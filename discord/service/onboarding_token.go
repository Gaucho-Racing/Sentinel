package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
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

// OnboardingConsumePayload is the form data the user submits at the end of
// the onboarding flow.
type OnboardingConsumePayload struct {
	Email                 string
	Password              string
	Username              string
	FirstName             string
	LastName              string
	Gender                string
	Birthday              string
	PhoneNumber           string
	GraduateLevel         string
	GraduationYear        int
	Major                 string
	ShirtSize             string
	JacketSize            string
	SAERegistrationNumber string
	InitialRole           string
}

// ConsumeOnboardingToken validates the token, fans out 5 calls to core to
// create Entity + User + email/phone/external auth, then marks the token used.
// Best-effort: a mid-flight failure leaves an orphan Entity that a cleanup
// job can sweep later. Returns the new entity ID.
func ConsumeOnboardingToken(id string, p OnboardingConsumePayload) (string, error) {
	token, err := GetOnboardingTokenByID(id)
	if err != nil {
		return "", err
	}

	birthday, err := time.Parse("2006-01-02", p.Birthday)
	if err != nil {
		return "", fmt.Errorf("invalid birthday format (want YYYY-MM-DD): %w", err)
	}

	var entityResp struct {
		ID string `json:"id"`
	}
	if err := sentinel.Post("/api/core/entity", map[string]string{"type": "USER"}, &entityResp); err != nil {
		return "", fmt.Errorf("create entity: %w", err)
	}

	var ignored map[string]any

	userBody := map[string]any{
		"entity_id":               entityResp.ID,
		"username":                p.Username,
		"first_name":              p.FirstName,
		"last_name":               p.LastName,
		"gender":                  p.Gender,
		"birthday":                birthday,
		"graduate_level":          p.GraduateLevel,
		"graduation_year":         p.GraduationYear,
		"major":                   p.Major,
		"shirt_size":              p.ShirtSize,
		"jacket_size":             p.JacketSize,
		"sae_registration_number": p.SAERegistrationNumber,
		"avatar_url":              token.DiscordAvatarURL,
		"initial_role":            p.InitialRole,
	}
	if err := sentinel.Post("/api/core/users", userBody, &ignored); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	if err := sentinel.Post(fmt.Sprintf("/core/entity/%s/email-auth", entityResp.ID), map[string]string{
		"email":    p.Email,
		"password": p.Password,
	}, &ignored); err != nil {
		return "", fmt.Errorf("create email auth: %w", err)
	}

	if err := sentinel.Post(fmt.Sprintf("/core/entity/%s/phone-auth", entityResp.ID), map[string]string{
		"phone_number": p.PhoneNumber,
	}, &ignored); err != nil {
		return "", fmt.Errorf("create phone auth: %w", err)
	}

	if err := sentinel.Post(fmt.Sprintf("/core/entity/%s/external-auth", entityResp.ID), map[string]string{
		"provider":    "DISCORD",
		"external_id": token.DiscordID,
	}, &ignored); err != nil {
		return "", fmt.Errorf("create external auth: %w", err)
	}

	now := time.Now()
	if err := database.DB.Model(&model.OnboardingToken{}).
		Where("id = ?", id).
		Updates(map[string]any{"used_at": now, "entity_id": entityResp.ID}).
		Error; err != nil {
		logger.SugarLogger.Errorf("Failed to mark onboarding token %s used: %v", id, err)
	}

	if err := AssignOnboardingRoles(token.DiscordID, p.InitialRole); err != nil {
		logger.SugarLogger.Errorf("Failed to assign onboarding roles for entity %s (discord_id=%s, initial_role=%s): %v", entityResp.ID, token.DiscordID, p.InitialRole, err)
	}

	return entityResp.ID, nil
}
