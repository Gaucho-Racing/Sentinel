package service

// GetEntityIDForDiscordUser resolves a Discord user ID to a Sentinel entity ID.
// Returns "" if no mapping is found, allowing callers to persist records
// with an empty entity_id that can be backfilled later.
func GetEntityIDForDiscordUser(discordUserID string) string {
	// TODO: implement lookup via EntityExternalAuth
	return ""
}
