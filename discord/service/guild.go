package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

func GetGuildRoles() ([]*discordgo.Role, error) {
	roles, err := Discord.GuildRoles(config.DiscordGuild)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Position < roles[j].Position
	})
	return roles, nil
}

func GetGuildChannels() ([]*discordgo.Channel, error) {
	channels, err := Discord.GuildChannels(config.DiscordGuild)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Position < channels[j].Position
	})
	return channels, nil
}

// GetGuildMember returns the member record for a user in the configured
// guild, preferring the in-memory State cache and falling back to a
// REST call. Returns an error if the user is not a member of the guild.
func GetGuildMember(userID string) (*discordgo.Member, error) {
	if m, err := Discord.State.Member(config.DiscordGuild, userID); err == nil && m != nil {
		return m, nil
	}
	return Discord.GuildMember(config.DiscordGuild, userID)
}

// DiscordRolesForInitialRole returns the guild role IDs that should be
// granted for a given onboarding initial_role value. Unknown values return
// nil so callers no-op rather than guess.
func DiscordRolesForInitialRole(initialRole string) []string {
	switch initialRole {
	case "member":
		return []string{config.MembersDiscordRoleID}
	case "alumni":
		return []string{config.AlumniDiscordRoleID}
	case "guest":
		return []string{config.GuestDiscordRoleID}
	default:
		return nil
	}
}

// SetGuildNickname sets the user's nickname in the configured guild.
// Empty or whitespace-only names are skipped. Discord's nickname field is
// capped at 32 characters; longer values are truncated.
func SetGuildNickname(discordID, nickname string) error {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return nil
	}
	if len(trimmed) > 32 {
		trimmed = trimmed[:32]
	}
	_, err := Discord.GuildMemberEdit(config.DiscordGuild, discordID, &discordgo.GuildMemberParams{Nick: trimmed})
	return err
}

// AssignOnboardingRoles grants the Discord roles mapped to a user's
// initial_role. Each grant is best-effort and logged individually so a
// single failure doesn't skip the remaining roles. Returns the first error
// encountered (if any) for the caller to surface, but does not stop on it.
func AssignOnboardingRoles(discordID, initialRole string) error {
	roleIDs := DiscordRolesForInitialRole(initialRole)
	if len(roleIDs) == 0 {
		return nil
	}
	var firstErr error
	for _, roleID := range roleIDs {
		if err := Discord.GuildMemberRoleAdd(config.DiscordGuild, discordID, roleID); err != nil {
			logger.SugarLogger.Errorf("Failed to add role %s to discord user %s (initial_role=%s): %v", roleID, discordID, initialRole, err)
			if firstErr == nil {
				firstErr = fmt.Errorf("add role %s: %w", roleID, err)
			}
		}
	}
	return firstErr
}
