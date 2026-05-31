package service

import (
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
)

type entityResponse struct {
	ID   string         `json:"id"`
	Type string         `json:"type"`
	User map[string]any `json:"user"`
}

// GetEntityIDForDiscordUser resolves a Discord user ID to a Sentinel entity ID.
// Returns "" if no mapping is found, allowing callers to persist records
// with an empty entity_id that can be backfilled later.
func GetEntityIDForDiscordUser(discordUserID string) string {
	var entity entityResponse
	err := sentinel.Get("/core/entity/external/DISCORD/"+discordUserID, &entity)
	if err != nil {
		logger.SugarLogger.Debugf("No entity found for Discord user %s: %v", discordUserID, err)
		return ""
	}
	return entity.ID
}

// SyncDiscordUserAvatar mirrors a Discord user's avatar onto the linked
// Sentinel user, when one exists. No-ops silently when the Discord user
// has no Sentinel record or the avatar is already current.
func SyncDiscordUserAvatar(discordUserID, avatarURL string) {
	var entity entityResponse
	if err := sentinel.Get("/core/entity/external/DISCORD/"+discordUserID, &entity); err != nil {
		logger.SugarLogger.Debugf("avatar sync: no entity for Discord user %s: %v", discordUserID, err)
		return
	}
	if entity.User == nil {
		return
	}
	if current, _ := entity.User["avatar_url"].(string); current == avatarURL {
		return
	}
	entity.User["avatar_url"] = avatarURL
	if err := sentinel.Post("/core/users", entity.User, nil); err != nil {
		logger.SugarLogger.Errorf("avatar sync: failed to update user %v: %v", entity.User["id"], err)
		return
	}
	logger.SugarLogger.Infof("avatar sync: updated user %v avatar to %s", entity.User["id"], avatarURL)
}
