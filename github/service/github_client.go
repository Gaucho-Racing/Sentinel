package service

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gaucho-racing/sentinel/github/config"
	"github.com/golang-jwt/jwt/v5"
)

type githubClient struct {
	config  config.Configuration
	key     *rsa.PrivateKey
	mu      sync.Mutex
	token   string
	expires time.Time
}

type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

func NewGitHubClient(cfg config.Configuration) (*githubClient, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse GitHub App private key: %w", err)
	}
	return &githubClient{config: cfg, key: key}, nil
}

func (g *githubClient) installationToken(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.token != "" && time.Now().Add(5*time.Minute).Before(g.expires) {
		return g.token, nil
	}
	now := time.Now()
	appJWT, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": g.config.AppID,
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
	}).SignedString(g.key)
	if err != nil {
		return "", err
	}
	var result struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	endpoint := "https://api.github.com/app/installations/" + strconv.FormatInt(g.config.InstallationID, 10) + "/access_tokens"
	if err := jsonRequest(ctx, http.MethodPost, endpoint, appJWT, map[string]any{}, &result); err != nil {
		return "", fmt.Errorf("create GitHub installation token: %w", err)
	}
	if result.Token == "" || result.ExpiresAt.IsZero() {
		return "", fmt.Errorf("GitHub returned an invalid installation token")
	}
	g.token, g.expires = result.Token, result.ExpiresAt
	return g.token, nil
}

func (g *githubClient) request(ctx context.Context, method, path string, input, output any) error {
	token, err := g.installationToken(ctx)
	if err != nil {
		return err
	}
	return jsonRequest(ctx, method, "https://api.github.com"+path, token, input, output)
}

func (g *githubClient) list(ctx context.Context, path string, output any) error {
	return g.request(ctx, http.MethodGet, path, nil, output)
}

func (g *githubClient) members(ctx context.Context) ([]githubUser, error) {
	var all []githubUser
	for page := 1; ; page++ {
		var batch []githubUser
		path := fmt.Sprintf("/orgs/%s/members?per_page=100&page=%d", url.PathEscape(g.config.Org), page)
		if err := g.list(ctx, path, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			return all, nil
		}
	}
}

type githubInvitation struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

func (g *githubClient) invitations(ctx context.Context) ([]githubInvitation, error) {
	var all []githubInvitation
	for page := 1; ; page++ {
		var batch []githubInvitation
		path := fmt.Sprintf("/orgs/%s/invitations?per_page=100&page=%d", url.PathEscape(g.config.Org), page)
		if err := g.list(ctx, path, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			return all, nil
		}
	}
}

func (g *githubClient) resolveUser(ctx context.Context, id int64) (githubUser, error) {
	var user githubUser
	err := g.list(ctx, "/user/"+strconv.FormatInt(id, 10), &user)
	return user, err
}

func (g *githubClient) membership(ctx context.Context, login string) (string, error) {
	var membership struct {
		Role  string `json:"role"`
		State string `json:"state"`
	}
	path := "/orgs/" + url.PathEscape(g.config.Org) + "/memberships/" + url.PathEscape(login)
	err := g.list(ctx, path, &membership)
	if apiErr, ok := err.(*httpError); ok && apiErr.status == http.StatusNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if membership.State != "active" {
		return "", nil
	}
	return membership.Role, nil
}

func (g *githubClient) invite(ctx context.Context, userID int64, role string) error {
	path := "/orgs/" + url.PathEscape(g.config.Org) + "/invitations"
	invitationRole := "direct_member"
	if role == "admin" {
		invitationRole = "admin"
	}
	return g.request(ctx, http.MethodPost, path, map[string]any{"invitee_id": userID, "role": invitationRole}, nil)
}

func (g *githubClient) setMembership(ctx context.Context, login, role string) error {
	path := "/orgs/" + url.PathEscape(g.config.Org) + "/memberships/" + url.PathEscape(login)
	return g.request(ctx, http.MethodPut, path, map[string]string{"role": role}, nil)
}

func (g *githubClient) removeMembership(ctx context.Context, login string) error {
	path := "/orgs/" + url.PathEscape(g.config.Org) + "/memberships/" + url.PathEscape(login)
	return g.request(ctx, http.MethodDelete, path, nil, nil)
}

func (g *githubClient) cancelInvitation(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/orgs/%s/invitations/%d", url.PathEscape(g.config.Org), id)
	return g.request(ctx, http.MethodDelete, path, nil, nil)
}

func (g *githubClient) authorizeURL(state string) string {
	params := url.Values{"client_id": {g.config.ClientID}, "redirect_uri": {g.config.RedirectURI}, "state": {state}}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (g *githubClient) exchangeCode(ctx context.Context, code string) (githubUser, error) {
	fields := url.Values{"client_id": {g.config.ClientID}, "client_secret": {g.config.ClientSecret}, "code": {code}, "redirect_uri": {g.config.RedirectURI}}
	var token struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := formRequest(ctx, "https://github.com/login/oauth/access_token", fields, &token); err != nil {
		return githubUser{}, err
	}
	if token.Error != "" || token.AccessToken == "" {
		return githubUser{}, fmt.Errorf("GitHub authorization failed: %s", token.Error)
	}
	var user githubUser
	if err := jsonRequest(ctx, http.MethodGet, "https://api.github.com/user", token.AccessToken, nil, &user); err != nil {
		return githubUser{}, err
	}
	if user.ID <= 0 || user.Login == "" {
		return githubUser{}, fmt.Errorf("GitHub returned an invalid account")
	}
	return user, nil
}
