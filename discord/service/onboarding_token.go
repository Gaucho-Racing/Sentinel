package service

import (
	"errors"
	"time"

	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

var (
	ErrOnboardingTokenNotFound = errors.New("onboarding token not found")
	ErrOnboardingTokenInvalid  = errors.New("onboarding token expired or already used")
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

// GetOnboardingTokenByID returns the token row if it exists and is still
// usable. ErrOnboardingTokenNotFound for a missing row, ErrOnboardingTokenInvalid
// for a row that is expired or already consumed.
func GetOnboardingTokenByID(id string) (model.OnboardingToken, error) {
	var token model.OnboardingToken
	if err := database.DB.Where("id = ?", id).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.OnboardingToken{}, ErrOnboardingTokenNotFound
		}
		return model.OnboardingToken{}, err
	}
	if token.UsedAt != nil || time.Now().After(token.ExpiresAt) {
		return token, ErrOnboardingTokenInvalid
	}
	return token, nil
}
