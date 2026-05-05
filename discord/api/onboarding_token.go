package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
	"github.com/gin-gonic/gin"
)

type onboardingTokenInfo struct {
	DiscordID         string `json:"discord_id"`
	DiscordUsername   string `json:"discord_username"`
	DiscordGlobalName string `json:"discord_global_name"`
	DiscordAvatarURL  string `json:"discord_avatar_url"`
}

func GetOnboardingToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id := c.Param("id")
	token, err := service.GetOnboardingTokenByID(id)
	switch {
	case errors.Is(err, service.ErrOnboardingTokenNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "onboarding token not found"})
		return
	case errors.Is(err, service.ErrOnboardingTokenInvalid):
		c.JSON(http.StatusGone, gin.H{"error": "onboarding token expired or already used"})
		return
	case err != nil:
		logger.SugarLogger.Errorf("Failed to fetch onboarding token %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, onboardingTokenInfo{
		DiscordID:         token.DiscordID,
		DiscordUsername:   token.DiscordUsername,
		DiscordGlobalName: token.DiscordGlobalName,
		DiscordAvatarURL:  token.DiscordAvatarURL,
	})
}
