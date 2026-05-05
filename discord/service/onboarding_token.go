package service

import (
	"time"

	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

// CreateOnboardingTokenForDiscordUser invalidates any unused tokens for the
// Discord user and mints a fresh one. Caller is expected to have already
// verified that no Entity exists for this Discord ID.
func CreateOnboardingTokenForDiscordUser(discordID, username, globalName, avatarURL string) (model.OnboardingToken, error) {
	now := time.Now()
	if err := database.DB.Model(&model.OnboardingToken{}).
		Where("discord_id = ? AND used_at IS NULL AND expires_at > ?", discordID, now).
		Update("expires_at", now).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to invalidate prior onboarding tokens for %s: %v", discordID, err)
	}

	token := model.OnboardingToken{
		ID:                ulid.Make().Prefixed("ont"),
		DiscordID:         discordID,
		DiscordUsername:   username,
		DiscordGlobalName: globalName,
		DiscordAvatarURL:  avatarURL,
		ExpiresAt:         now.Add(config.OnboardingTokenTTL),
	}
	if err := database.DB.Create(&token).Error; err != nil {
		return model.OnboardingToken{}, err
	}
	return token, nil
}
