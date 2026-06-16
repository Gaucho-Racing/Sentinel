package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

// LoginDiscord completes the "Continue with Discord" flow:
//
//  1. Exchange the OAuth code (handed back to the web client by Discord's
//     redirect) for a Discord access token.
//  2. Read the authenticated user's Discord ID.
//  3. Look up the Sentinel entity linked to that Discord ID via core.
//  4. If found, mint a first-party Sentinel session (same shape as
//     email-password login). If not found, return a structured 404 the web
//     can branch on to show the "run !verify in Discord" page.
//
// The web client posts the code as a query param to mirror the v4 contract.
func LoginDiscord(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := service.ExchangeDiscordCode(code)
	if err != nil {
		logger.SugarLogger.Errorf("discord login: code exchange failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired code"})
		return
	}

	user, err := service.GetDiscordUser(token.AccessToken)
	if err != nil {
		logger.SugarLogger.Errorf("discord login: user lookup failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load discord user"})
		return
	}

	var entity struct {
		ID string `json:"id"`
	}
	if err := sentinel.Get("/api/core/entity/external/DISCORD/"+user.ID, &entity); err != nil {
		var apiErr *sentinel.APIError
		// Upstream 404 = no Sentinel entity is linked to this Discord ID.
		// Distinct error code so the web can route to the !verify-instructions
		// page instead of a generic failure toast.
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "no_account",
				"message": "No Sentinel account is linked to this Discord user.",
			})
			return
		}
		logger.SugarLogger.Errorf("discord login: entity lookup failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Best-effort: refresh the provider metadata so the latest email /
	// username / avatar we can see from Discord is on the row. Failures here
	// shouldn't block the login — the session matters more than the audit.
	metadata := map[string]any{
		"email":       user.Email,
		"username":    user.Username,
		"global_name": user.GlobalName,
		"avatar":      user.Avatar,
		"verified":    user.Verified,
	}
	if err := sentinel.Patch(
		"/api/core/entity/"+entity.ID+"/external-auth/DISCORD",
		map[string]any{"metadata": metadata},
		nil,
	); err != nil {
		logger.SugarLogger.Warnf("discord login: metadata refresh failed for entity %s: %v", entity.ID, err)
	}

	resp, err := mintFirstPartySession(c, entity.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
