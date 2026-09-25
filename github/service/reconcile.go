package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gaucho-racing/sentinel/github/config"
	"github.com/gaucho-racing/sentinel/github/pkg/logger"
	"github.com/gaucho-racing/sentinel/github/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

type pendingLink struct {
	entityID string
	expires  time.Time
}

const teamAdminAccountID int64 = 153126024

type Server struct {
	config config.Configuration
	github *githubClient
	states sync.Map
	mu     sync.Mutex
}

func (s *Server) consumeLinkState(state string) (pendingLink, bool) {
	value, ok := s.states.LoadAndDelete(state)
	if !ok {
		return pendingLink{}, false
	}
	pending := value.(pendingLink)
	return pending, time.Now().Before(pending.expires)
}

func NewServer(cfg config.Configuration, github *githubClient) *Server {
	return &Server{config: cfg, github: github}
}

func (s *Server) BeginLink(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sign in to link your GitHub account"})
		return
	}
	entityID, err := verifyUser(c.Request.Context(), parts[1])
	if err != nil {
		if apiErr, ok := err.(*sentinel.APIError); ok && apiErr.Status == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "valid Sentinel user session required"})
		} else {
			logger.SugarLogger.Errorf("Sentinel session validation failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "could not verify Sentinel session"})
		}
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not begin linking"})
		return
	}
	state := base64.RawURLEncoding.EncodeToString(raw)
	s.states.Store(state, pendingLink{entityID: entityID, expires: time.Now().Add(10 * time.Minute)})
	time.AfterFunc(10*time.Minute, func() {
		if value, ok := s.states.Load(state); ok && time.Now().After(value.(pendingLink).expires) {
			s.states.Delete(state)
		}
	})
	c.JSON(http.StatusOK, gin.H{"url": s.github.authorizeURL(state)})
}

func (s *Server) FinishLink(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.String(http.StatusBadRequest, "invalid linking state")
		return
	}
	pending, ok := s.consumeLinkState(state)
	if !ok {
		c.String(http.StatusBadRequest, "link expired or already used")
		return
	}
	if c.Query("code") == "" {
		c.String(http.StatusBadRequest, "link expired or declined")
		return
	}
	user, err := s.github.exchangeCode(c.Request.Context(), c.Query("code"))
	if err != nil {
		logger.SugarLogger.Errorf("GitHub linking failed: %v", err)
		c.String(http.StatusBadGateway, "could not verify GitHub account")
		return
	}
	if err := linkIdentity(c.Request.Context(), pending.entityID, user.ID, user.Login); err != nil {
		logger.SugarLogger.Errorf("GitHub linking persistence failed: %v", err)
		if apiErr, ok := err.(*sentinel.APIError); ok && apiErr.Status == http.StatusConflict {
			c.String(http.StatusConflict, "GitHub account is already linked")
		} else {
			c.String(http.StatusBadGateway, "could not save GitHub link; try again")
		}
		return
	}
	query := url.Values{"github": {"linked"}}
	go s.RunReconcile(context.Background())
	c.Redirect(http.StatusSeeOther, "/profile?"+query.Encode())
}

func (s *Server) RunReconcile(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()
	if err := s.reconcile(ctx); err != nil {
		logger.SugarLogger.Errorf("GitHub org reconciliation failed: %v", err)
	}
}

func desiredRole(groups []sentinelGroup) string {
	member := false
	for _, group := range groups {
		switch group.Name {
		case "GithubAdmins":
			return "admin"
		case "GithubMembers":
			member = true
		}
	}
	if member {
		return "member"
	}
	return ""
}

func (s *Server) reconcile(ctx context.Context) error {
	links, err := identities(ctx)
	if err != nil {
		return fmt.Errorf("list Sentinel GitHub identities: %w", err)
	}
	members, err := s.github.members(ctx)
	if err != nil {
		return fmt.Errorf("list GitHub org members: %w", err)
	}
	invitations, err := s.github.invitations(ctx)
	if err != nil {
		return fmt.Errorf("list GitHub org invitations: %w", err)
	}
	roles := make(map[int64]string, len(links))
	logins := make(map[int64]string, len(links))
	for _, link := range links {
		if err := ctx.Err(); err != nil {
			return err
		}
		userID, err := strconv.ParseInt(link.ExternalID, 10, 64)
		if err != nil || userID <= 0 {
			return fmt.Errorf("invalid linked GitHub ID for %s", link.EntityID)
		}
		groups, err := groupsForEntity(ctx, link.EntityID)
		if err != nil {
			return fmt.Errorf("load groups for %s: %w", link.EntityID, err)
		}
		actual, err := s.github.resolveUser(ctx, userID)
		if err != nil {
			return fmt.Errorf("resolve GitHub user %d: %w", userID, err)
		}
		if actual.ID != userID || actual.Login == "" {
			return fmt.Errorf("GitHub account %d returned invalid identity", userID)
		}
		if _, duplicate := roles[userID]; duplicate {
			return fmt.Errorf("GitHub account %d linked to multiple Sentinel entities", userID)
		}
		roles[userID], logins[userID] = desiredRole(groups), actual.Login
	}
	current := make(map[int64]githubUser, len(members))
	for _, member := range members {
		if member.ID <= 0 || member.Login == "" {
			return fmt.Errorf("incomplete GitHub organization membership response")
		}
		if _, duplicate := current[member.ID]; duplicate {
			return fmt.Errorf("duplicate GitHub org member %d in response", member.ID)
		}
		current[member.ID] = member
	}
	teamAdmin, ok := current[teamAdminAccountID]
	if !ok {
		admin, err := s.github.resolveUser(ctx, teamAdminAccountID)
		if err != nil {
			return fmt.Errorf("resolve team admin account: %w", err)
		}
		if admin.ID != teamAdminAccountID || !strings.EqualFold(admin.Login, "gauchoracing") {
			return fmt.Errorf("team admin account identity mismatch; skipping reconciliation")
		}
		for _, invitation := range invitations {
			if strings.EqualFold(invitation.Login, admin.Login) {
				return fmt.Errorf("team admin account gauchoracing has a pending org invitation; waiting for acceptance")
			}
		}
		if err := s.github.invite(ctx, teamAdminAccountID, "admin"); err != nil {
			return fmt.Errorf("invite team admin account gauchoracing as owner: %w", err)
		}
		return fmt.Errorf("team admin account gauchoracing invited as owner; waiting for acceptance")
	}
	verifiedAdmin, err := s.github.resolveUser(ctx, teamAdminAccountID)
	if err != nil {
		return fmt.Errorf("verify team admin account: %w", err)
	}
	if verifiedAdmin.ID != teamAdminAccountID || !strings.EqualFold(verifiedAdmin.Login, teamAdmin.Login) || !strings.EqualFold(teamAdmin.Login, "gauchoracing") {
		return fmt.Errorf("team admin account ID mismatch; skipping reconciliation")
	}
	adminRole, err := s.github.membership(ctx, teamAdmin.Login)
	if err != nil {
		return fmt.Errorf("verify team admin owner role: %w", err)
	}
	if adminRole != "admin" {
		if err := s.github.setMembership(ctx, teamAdmin.Login, "admin"); err != nil {
			return fmt.Errorf("restore team admin owner role: %w", err)
		}
		logger.SugarLogger.Infof("restored GitHub org owner role for gauchoracing")
	}
	if err := validatePolicyGroups(ctx); err != nil {
		return fmt.Errorf("validate GitHub access groups: %w", err)
	}
	pendingLogins := make(map[string]struct{}, len(invitations))
	for _, invitation := range invitations {
		if invitation.ID <= 0 {
			return fmt.Errorf("incomplete GitHub invitation response")
		}
		if invitation.Login != "" {
			pendingLogins[strings.ToLower(invitation.Login)] = struct{}{}
		}
	}
	for id, role := range roles {
		if err := ctx.Err(); err != nil {
			return err
		}
		if id == teamAdminAccountID {
			continue
		}
		if role == "" {
			continue
		}
		login := logins[id]
		present := current[id]
		if present.ID == 0 {
			if _, pending := pendingLogins[strings.ToLower(login)]; pending {
				continue
			}
			if err := s.github.invite(ctx, id, role); err != nil {
				logger.SugarLogger.Errorf("GitHub invite %s failed: %v", login, err)
			}
			continue
		}
		currentRole, err := s.github.membership(ctx, login)
		if err != nil {
			logger.SugarLogger.Errorf("GitHub membership read for %s failed: %v", login, err)
			continue
		}
		if currentRole != role {
			if err := s.github.setMembership(ctx, login, role); err != nil {
				logger.SugarLogger.Errorf("GitHub role update for %s failed: %v", login, err)
			}
		}
	}
	for _, member := range members {
		if member.ID == teamAdminAccountID || roles[member.ID] != "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		latestRole, err := s.github.membership(ctx, member.Login)
		if err != nil || latestRole == "" {
			if err != nil {
				logger.SugarLogger.Errorf("GitHub membership check for %s failed: %v", member.Login, err)
			}
			continue
		}
		if err := s.github.removeMembership(ctx, member.Login); err != nil {
			logger.SugarLogger.Errorf("GitHub org removal for %s failed: %v", member.Login, err)
		}
	}
	for _, invitation := range invitations {
		if err := ctx.Err(); err != nil {
			return err
		}
		if invitation.Login == "" {
			continue
		}
		keep := strings.EqualFold(invitation.Login, teamAdmin.Login)
		for id, login := range logins {
			if strings.EqualFold(login, invitation.Login) && roles[id] != "" {
				keep = true
				break
			}
		}
		if !keep {
			if err := s.github.cancelInvitation(ctx, invitation.ID); err != nil {
				logger.SugarLogger.Errorf("GitHub invitation cancellation for %s failed: %v", invitation.Login, err)
			}
		}
	}
	return nil
}
