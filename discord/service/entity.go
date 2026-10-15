package service

import (
	"fmt"
	"maps"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
)

type entityResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	User      map[string]any `json:"user"`
	EmailAuth struct {
		Email string `json:"email"`
	} `json:"email_auth"`
	ExternalAuths []struct {
		Provider string         `json:"provider"`
		Metadata map[string]any `json:"metadata"`
	} `json:"external_auths"`
}

// GetEntityIDForDiscordUser resolves a Discord user ID to a Sentinel entity ID.
// Returns "" if no mapping is found, allowing callers to persist records
// with an empty entity_id that can be backfilled later.
func GetEntityIDForDiscordUser(discordUserID string) string {
	var entity entityResponse
	err := sentinel.Get("/api/core/entity/external/DISCORD/"+discordUserID, &entity)
	if err != nil {
		logger.SugarLogger.Debugf("No entity found for Discord user %s: %v", discordUserID, err)
		return ""
	}
	return entity.ID
}

// GetEntityEmailForDiscordUser resolves a Discord user ID to the linked
// Sentinel entity's primary login email. Returns "" if no entity is linked
// or the entity has no email auth row. Used by !verify to prefill the
// already-onboarded login link.
func GetEntityEmailForDiscordUser(discordUserID string) string {
	var entity entityResponse
	if err := sentinel.Get("/api/core/entity/external/DISCORD/"+discordUserID, &entity); err != nil {
		logger.SugarLogger.Debugf("No entity found for Discord user %s: %v", discordUserID, err)
		return ""
	}
	return entity.EmailAuth.Email
}

type groupResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetGroupNamesForDiscordUser returns the names of the Sentinel groups the
// Discord user's linked entity belongs to. Returns an error when the user
// has no linked entity or the core lookup fails, so authorization callers
// can fail closed.
func GetGroupNamesForDiscordUser(discordUserID string) ([]string, error) {
	entityID := GetEntityIDForDiscordUser(discordUserID)
	if entityID == "" {
		return nil, fmt.Errorf("no sentinel entity linked to discord user %s", discordUserID)
	}
	var groups []groupResponse
	if err := sentinel.Get("/api/core/entity/"+entityID+"/groups", &groups); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group.Name)
	}
	return names, nil
}

// SyncDiscordUserProfile mirrors a Discord user's avatar onto the linked
// Sentinel user and their username onto the linked Discord identity. No-ops
// when the Discord user has no Sentinel record or nothing changed.
func SyncDiscordUserProfile(user *discordgo.User, avatarURL string) {
	var entity entityResponse
	if err := sentinel.Get("/api/core/entity/external/DISCORD/"+user.ID, &entity); err != nil {
		logger.SugarLogger.Debugf("profile sync: no entity for Discord user %s: %v", user.ID, err)
		return
	}
	for _, auth := range entity.ExternalAuths {
		if auth.Provider == "DISCORD" {
			SyncDiscordIdentityMetadata(entity.ID, auth.Metadata, user)
		}
	}
	if entity.User == nil {
		return
	}
	if current, _ := entity.User["avatar_url"].(string); current == avatarURL {
		return
	}
	entity.User["avatar_url"] = avatarURL
	if err := sentinel.Post("/api/core/users", entity.User, nil); err != nil {
		logger.SugarLogger.Errorf("avatar sync: failed to update user %v: %v", entity.User["id"], err)
		return
	}
	logger.SugarLogger.Infof("avatar sync: updated user %v avatar to %s", entity.User["id"], avatarURL)
}

// SyncDiscordIdentityMetadata stores the Discord username and display name on
// the entity's Discord identity, keeping whatever else the metadata holds.
func SyncDiscordIdentityMetadata(entityID string, metadata map[string]any, user *discordgo.User) {
	if user == nil || user.Username == "" {
		return
	}
	currentUsername, _ := metadata["username"].(string)
	currentGlobalName, _ := metadata["global_name"].(string)
	if currentUsername == user.Username && currentGlobalName == user.GlobalName {
		return
	}
	updated := maps.Clone(metadata)
	if updated == nil {
		updated = map[string]any{}
	}
	updated["username"] = user.Username
	updated["global_name"] = user.GlobalName
	if err := sentinel.Patch("/api/core/entity/"+entityID+"/external-auth/DISCORD", map[string]any{"metadata": updated}, nil); err != nil {
		logger.SugarLogger.Errorf("identity sync: failed to update Discord username for entity %s: %v", entityID, err)
		return
	}
	logger.SugarLogger.Infof("identity sync: entity %s Discord username is now %s", entityID, user.Username)
}
