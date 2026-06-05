package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

// writeGateError translates a CheckAccessGate failure into a response. A
// genuine denial is 403 access_denied; any other error means the gate couldn't
// be evaluated (a core fetch failed) — we fail closed with 502 rather than let
// the login through.
func writeGateError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrAccessDenied) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "error_description": err.Error()})
		return
	}
	logger.SugarLogger.Errorf("access gate evaluation failed: %v", err)
	c.JSON(http.StatusBadGateway, gin.H{"error": "server_error", "error_description": "could not verify access"})
}

type applicationResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ClientID     string   `json:"client_id"`
	IconURL      string   `json:"icon_url"`
	RedirectURIs []string `json:"redirect_uris"`
}

type validateAuthorizeResponse struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	Scope       string `json:"scope"`
	Prompt      string `json:"prompt"`
	AppName     string `json:"app_name"`
	AppIconURL  string `json:"app_icon_url"`
}

// ValidateAuthorize validates the OAuth authorize request parameters
// and returns application info for the frontend consent screen.
func ValidateAuthorize(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri is required"})
		return
	}

	scope := c.Query("scope")
	if scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope is required"})
		return
	}

	// The authorization code flow is the only response type we support.
	if responseType := c.Query("response_type"); responseType != "" && responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_response_type"})
		return
	}

	if !service.ValidateScopes(scope) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}

	if service.ScopesContain(scope, "sentinel:all") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sentinel:all scope cannot be requested by client applications"})
		return
	}

	var app applicationResponse
	err := sentinel.Get("/api/applications/client/"+clientID, &app)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get application for client_id %s: %v", clientID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	validURI := false
	for _, uri := range app.RedirectURIs {
		if service.MatchRedirectURI(uri, redirectURI) {
			validURI = true
			break
		}
	}
	if !validURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect_uri"})
		return
	}

	// Enforce the access gate here (not only at the authorize/token steps) so a
	// user who doesn't qualify gets a clear error page up front, instead of a
	// consent screen followed by a redirect back to the client with
	// access_denied. entity_id is supplied by the SPA from the active session.
	entityID := c.Query("entity_id")
	if entityID != "" {
		if err := service.CheckAccessGate(entityID, clientID); err != nil {
			if errors.Is(err, service.ErrAccessDenied) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "app_name": app.Name, "app_icon_url": app.IconURL})
				return
			}
			logger.SugarLogger.Errorf("access gate evaluation failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "server_error"})
			return
		}
	}

	// Default to a consent prompt. If the user already authorized this exact
	// client+scope set within the last 24h, skip the screen and auto-approve.
	prompt := "consent"
	if entityID != "" {
		q := url.Values{}
		q.Set("client_id", clientID)
		q.Set("scope", scope)
		q.Set("after", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
		q.Set("limit", "1")
		var logins []map[string]interface{}
		if err := sentinel.Get(fmt.Sprintf("/api/core/entity/%s/logins?%s", entityID, q.Encode()), &logins); err == nil && len(logins) > 0 {
			prompt = "none"
		}
	}

	c.JSON(http.StatusOK, validateAuthorizeResponse{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		Prompt:      prompt,
		AppName:     app.Name,
		AppIconURL:  app.IconURL,
	})
}

type authorizeRequest struct {
	EntityID string `json:"entity_id" binding:"required"`
}

// Authorize generates an authorization code after the user approves consent.
// The frontend sends the entity_id of the authenticated user.
func Authorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")

	if clientID == "" || redirectURI == "" || scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id, redirect_uri, and scope are required"})
		return
	}

	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !service.ValidateScopes(scope) || service.ScopesContain(scope, "sentinel:all") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}

	if err := service.CheckAccessGate(req.EntityID, clientID); err != nil {
		writeGateError(c, err)
		return
	}

	authCode, err := service.GenerateAuthorizationCode(req.EntityID, clientID, scope, redirectURI, c.Query("nonce"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         authCode.Code,
		"redirect_uri": redirectURI,
	})
}
