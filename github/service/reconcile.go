package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
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

func authenticatedEntityID(c *gin.Context) (string, bool) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sign in to manage your GitHub account"})
		return "", false
	}
	entityID, err := verifyUser(c.Request.Context(), parts[1])
	if err != nil {
		if apiErr, ok := err.(*sentinel.APIError); ok && apiErr.Status == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "valid Sentinel user session required"})
		} else {
			logger.SugarLogger.Errorf("Sentinel session validation failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "could not verify Sentinel session"})
		}
		return "", false
	}
	return entityID, true
}

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
	entityID, ok := authenticatedEntityID(c)
	if !ok {
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
	s.mu.Lock()
	err = linkIdentity(c.Request.Context(), pending.entityID, user.ID, user.Login)
	s.mu.Unlock()
	if err != nil {
		logger.SugarLogger.Errorf("GitHub linking persistence failed: %v", err)
		if apiErr, ok := err.(*sentinel.APIError); ok && apiErr.Status == http.StatusConflict {
			c.String(http.StatusConflict, "GitHub account is already linked")
		} else {
			c.String(http.StatusBadGateway, "could not save GitHub link; try again")
		}
		return
	}
	query := url.Values{"github": {"linked"}}
	logger.SugarLogger.Infof("GitHub org sync: account linked for entity %s, scheduling account reconcile", pending.entityID)
	go s.RunAccountReconcile(context.Background(), pending.entityID, user.ID)
	c.Redirect(http.StatusSeeOther, "/settings/connected-accounts?"+query.Encode())
}

func (s *Server) Unlink(c *gin.Context) {
	entityID, ok := authenticatedEntityID(c)
	if !ok {
		return
	}
	linked, err := linkedGitHubIdentity(c.Request.Context(), entityID)
	if err != nil {
		if errors.Is(err, errGitHubLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub account not linked"})
			return
		}
		logger.SugarLogger.Errorf("GitHub unlink lookup for entity %s failed: %v", entityID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load GitHub link"})
		return
	}
	id, err := strconv.ParseInt(linked.ExternalID, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid GitHub link"})
		return
	}
	if id == teamAdminAccountID {
		c.JSON(http.StatusForbidden, gin.H{"error": "the team admin account cannot be unlinked"})
		return
	}
	s.mu.Lock()
	_, err = unlinkIdentity(c.Request.Context(), entityID, linked.ExternalID)
	s.mu.Unlock()
	if err != nil {
		logger.SugarLogger.Errorf("GitHub unlink for entity %s failed: %v", entityID, err)
		var apiErr *sentinel.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "GitHub account changed; refresh and try again"})
		} else {
			c.JSON(http.StatusBadGateway, gin.H{"error": "could not unlink GitHub account"})
		}
		return
	}
	logger.SugarLogger.Infof("GitHub org sync: account unlinked for entity %s, scheduling account cleanup", entityID)
	go s.RunUnlinkedAccountReconcile(context.Background(), entityID, id)
	c.Status(http.StatusNoContent)
}

func (s *Server) LinkStatus(c *gin.Context) {
	entityID, ok := authenticatedEntityID(c)
	if !ok {
		return
	}
	linked, err := linkedGitHubIdentity(c.Request.Context(), entityID)
	if err != nil {
		if errors.Is(err, errGitHubLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub account not linked"})
			return
		}
		logger.SugarLogger.Errorf("GitHub status lookup for entity %s failed: %v", entityID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load GitHub link"})
		return
	}
	id, err := strconv.ParseInt(linked.ExternalID, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid GitHub link"})
		return
	}
	user, err := s.github.resolveUser(c.Request.Context(), id)
	if err != nil || user.ID != id || user.Login == "" {
		logger.SugarLogger.Errorf("GitHub status failed to resolve linked account %d: %v", id, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load GitHub account"})
		return
	}
	role, err := s.github.membership(c.Request.Context(), user.Login)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not check GitHub membership"})
		return
	}
	status := "not_invited"
	if role != "" {
		status = "active"
	} else {
		invitations, err := s.github.invitations(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "could not check GitHub invitation"})
			return
		}
		for _, invitation := range invitations {
			if strings.EqualFold(invitation.Login, user.Login) {
				status = "pending"
				break
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": status, "username": user.Login})
}

func (s *Server) RunReconcile(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()
	started := time.Now()
	logger.SugarLogger.Infof("GitHub org sync: starting full sweep for %s", s.config.Org)
	if err := s.reconcile(ctx); err != nil {
		logger.SugarLogger.Errorf("GitHub org sync: full sweep failed after %s: %v", time.Since(started).Round(time.Millisecond), err)
		return
	}
	logger.SugarLogger.Infof("GitHub org sync: full sweep complete in %s", time.Since(started).Round(time.Millisecond))
}

func (s *Server) RunAccountReconcile(ctx context.Context, entityID string, githubID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	started := time.Now()
	logger.SugarLogger.Infof("GitHub org sync: starting account reconcile for entity %s (GitHub ID %d)", entityID, githubID)
	if err := s.reconcileAccount(ctx, entityID, githubID); err != nil {
		logger.SugarLogger.Errorf("GitHub org sync: account reconcile failed after %s for entity %s: %v", time.Since(started).Round(time.Millisecond), entityID, err)
		return
	}
	logger.SugarLogger.Infof("GitHub org sync: account reconcile complete in %s for entity %s", time.Since(started).Round(time.Millisecond), entityID)
}

func (s *Server) RunUnlinkedAccountReconcile(ctx context.Context, entityID string, githubID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	started := time.Now()
	logger.SugarLogger.Infof("GitHub org sync: starting unlinked account cleanup for entity %s (GitHub ID %d)", entityID, githubID)
	if err := s.reconcileUnlinkedAccount(ctx, githubID); err != nil {
		logger.SugarLogger.Errorf("GitHub org sync: unlinked account cleanup failed after %s for entity %s: %v", time.Since(started).Round(time.Millisecond), entityID, err)
		return
	}
	logger.SugarLogger.Infof("GitHub org sync: unlinked account cleanup complete in %s for entity %s", time.Since(started).Round(time.Millisecond), entityID)
}

func (s *Server) recheckUnlinked(ctx context.Context, githubID int64) error {
	linked, err := identityForGitHubID(ctx, githubID)
	if err != nil {
		return fmt.Errorf("check current GitHub link: %w", err)
	}
	if linked {
		return errGitHubLinkChanged
	}
	return nil
}

var errGitHubLinkChanged = errors.New("GitHub account has been linked again")

func (s *Server) reconcileUnlinkedAccount(ctx context.Context, githubID int64) error {
	if githubID == teamAdminAccountID {
		return fmt.Errorf("team admin account cannot be removed")
	}
	if err := s.recheckUnlinked(ctx, githubID); errors.Is(err, errGitHubLinkChanged) {
		logger.SugarLogger.Infof("GitHub org sync: account %d was relinked; skipping cleanup", githubID)
		return nil
	} else if err != nil {
		return err
	}
	user, err := s.github.resolveUser(ctx, githubID)
	if err != nil {
		return fmt.Errorf("resolve GitHub account: %w", err)
	}
	if user.ID != githubID || user.Login == "" {
		return fmt.Errorf("invalid GitHub account identity")
	}
	membership, err := s.github.membership(ctx, user.Login)
	if err != nil {
		return fmt.Errorf("read GitHub membership: %w", err)
	}
	if membership != "" {
		if err := s.recheckUnlinked(ctx, githubID); errors.Is(err, errGitHubLinkChanged) {
			logger.SugarLogger.Infof("GitHub org sync: account %s was relinked; skipping cleanup", user.Login)
			return nil
		} else if err != nil {
			return err
		}
		if err := s.github.removeMembership(ctx, user.Login); err != nil {
			return fmt.Errorf("remove %s from org: %w", user.Login, err)
		}
		logger.SugarLogger.Infof("GitHub org sync: removed unlinked member %s", user.Login)
	}
	if err := s.recheckUnlinked(ctx, githubID); errors.Is(err, errGitHubLinkChanged) {
		logger.SugarLogger.Infof("GitHub org sync: account %s was relinked; skipping invitation cleanup", user.Login)
		return nil
	} else if err != nil {
		return err
	}
	invitations, err := s.github.invitations(ctx)
	if err != nil {
		return fmt.Errorf("read pending invitations: %w", err)
	}
	for _, invitation := range invitations {
		if strings.EqualFold(invitation.Login, user.Login) {
			if err := s.recheckUnlinked(ctx, githubID); errors.Is(err, errGitHubLinkChanged) {
				logger.SugarLogger.Infof("GitHub org sync: account %s was relinked; keeping invitation", user.Login)
				return nil
			} else if err != nil {
				return err
			}
			if err := s.github.cancelInvitation(ctx, invitation.ID); err != nil {
				return fmt.Errorf("cancel invitation for %s: %w", user.Login, err)
			}
			logger.SugarLogger.Infof("GitHub org sync: cancelled invitation for unlinked account %s", user.Login)
		}
	}
	return nil
}

func (s *Server) reconcileAccount(ctx context.Context, entityID string, githubID int64) error {
	if githubID == teamAdminAccountID {
		return fmt.Errorf("team admin account is managed by the full sweep")
	}
	user, err := s.github.resolveUser(ctx, githubID)
	if err != nil {
		return fmt.Errorf("resolve GitHub account: %w", err)
	}
	if user.ID != githubID || user.Login == "" {
		return fmt.Errorf("invalid GitHub account identity")
	}

	linked, err := linkedGitHubIdentity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("read current Sentinel GitHub link: %w", err)
	}
	if linked.ExternalID != strconv.FormatInt(githubID, 10) {
		return fmt.Errorf("GitHub link changed while reconciling")
	}
	groups, err := groupsForEntity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("read Sentinel groups: %w", err)
	}
	role := desiredRole(groups)
	currentRole, err := s.github.membership(ctx, user.Login)
	if err != nil {
		return fmt.Errorf("read GitHub membership: %w", err)
	}
	linkedNow, err := linkedGitHubIdentity(ctx, entityID)
	if err != nil {
		if errors.Is(err, errGitHubLinkNotFound) {
			logger.SugarLogger.Infof("GitHub org sync: entity %s was unlinked; skipping account grant", entityID)
			return nil
		}
		return fmt.Errorf("recheck GitHub link: %w", err)
	}
	if linkedNow.ExternalID != strconv.FormatInt(githubID, 10) {
		logger.SugarLogger.Infof("GitHub org sync: entity %s changed GitHub account; skipping account grant", entityID)
		return nil
	}
	if role == "" {
		if currentRole != "" {
			if err := s.github.removeMembership(ctx, user.Login); err != nil {
				return fmt.Errorf("remove ineligible account %s: %w", user.Login, err)
			}
			logger.SugarLogger.Infof("GitHub org sync: removed ineligible member %s", user.Login)
		} else {
			logger.SugarLogger.Infof("GitHub org sync: linked account %s has no GitHub access group", user.Login)
		}
		return nil
	}
	if currentRole == "" {
		invitations, err := s.github.invitations(ctx)
		if err != nil {
			return fmt.Errorf("read pending invitations: %w", err)
		}
		for _, invitation := range invitations {
			if strings.EqualFold(invitation.Login, user.Login) {
				logger.SugarLogger.Infof("GitHub org sync: invitation already pending for %s", user.Login)
				return nil
			}
		}
		if err := s.github.invite(ctx, githubID, role); err != nil {
			return fmt.Errorf("invite %s: %w", user.Login, err)
		}
		logger.SugarLogger.Infof("GitHub org sync: invited %s as %s", user.Login, role)
		return nil
	}
	if currentRole != role {
		if err := s.github.setMembership(ctx, user.Login, role); err != nil {
			return fmt.Errorf("update %s role: %w", user.Login, err)
		}
		logger.SugarLogger.Infof("GitHub org sync: updated %s role from %s to %s", user.Login, currentRole, role)
	}
	return nil
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
	logger.SugarLogger.Infof("GitHub org sync: fetched %d linked identities, %d org members, %d pending invitations", len(links), len(members), len(invitations))
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
	invited, rolesUpdated, removed, invitationsCancelled, failures := 0, 0, 0, 0, 0
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
				failures++
			} else {
				invited++
				logger.SugarLogger.Infof("GitHub org sync: invited %s as %s", login, role)
			}
			continue
		}
		currentRole, err := s.github.membership(ctx, login)
		if err != nil {
			logger.SugarLogger.Errorf("GitHub membership read for %s failed: %v", login, err)
			failures++
			continue
		}
		if currentRole != role {
			if err := s.github.setMembership(ctx, login, role); err != nil {
				logger.SugarLogger.Errorf("GitHub role update for %s failed: %v", login, err)
				failures++
			} else {
				rolesUpdated++
				logger.SugarLogger.Infof("GitHub org sync: updated %s role from %s to %s", login, currentRole, role)
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
				failures++
			}
			continue
		}
		if err := s.github.removeMembership(ctx, member.Login); err != nil {
			logger.SugarLogger.Errorf("GitHub org removal for %s failed: %v", member.Login, err)
			failures++
		} else {
			removed++
			logger.SugarLogger.Infof("GitHub org sync: removed unmatched member %s", member.Login)
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
				failures++
			} else {
				invitationsCancelled++
				logger.SugarLogger.Infof("GitHub org sync: cancelled unmatched invitation for %s", invitation.Login)
			}
		}
	}
	logger.SugarLogger.Infof("GitHub org sync: results invited=%d roles_updated=%d members_removed=%d invitations_cancelled=%d failures=%d", invited, rolesUpdated, removed, invitationsCancelled, failures)
	if failures > 0 {
		return fmt.Errorf("%d GitHub org actions failed", failures)
	}
	return nil
}
