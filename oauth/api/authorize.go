package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

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

	if !service.ValidateScopes(scope) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}

	if service.ScopesContain(scope, "sentinel:all") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sentinel:all scope cannot be requested by client applications"})
		return
	}

	var app applicationResponse
	err := sentinel.Get("/applications/client/"+clientID, &app)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get application for client_id %s: %v", clientID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	validURI := false
	for _, uri := range app.RedirectURIs {
		if uri == redirectURI {
			validURI = true
			break
		}
	}
	if !validURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect_uri"})
		return
	}

	prompt := c.Query("prompt")
	entityID := c.Query("entity_id")
	if prompt == "none" && entityID != "" {
		var logins []map[string]interface{}
		err = sentinel.Get(fmt.Sprintf("/core/entity/%s/logins?client_id=%s&scope=%s&limit=1", entityID, clientID, scope), &logins)
		if err == nil && len(logins) > 0 {
			if createdAt, ok := logins[0]["created_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, createdAt); err == nil && time.Since(t) < 7*24*time.Hour {
					prompt = "none"
				} else {
					prompt = "consent"
				}
			} else {
				prompt = "consent"
			}
		} else {
			prompt = "consent"
		}
	} else {
		prompt = "consent"
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

	authCode, err := service.GenerateAuthorizationCode(req.EntityID, clientID, scope, redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         authCode.Code,
		"redirect_uri": redirectURI,
	})
}
