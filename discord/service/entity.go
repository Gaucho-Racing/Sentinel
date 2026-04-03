package service

import (
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
)

type entityResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
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
