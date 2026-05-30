package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

func OnGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	logger.SugarLogger.Infof("GuildMemberAdd: user=%s roles=%v", m.User.ID, m.Roles)
}

func OnGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	if m.BeforeUpdate == nil {
		logger.SugarLogger.Infof("GuildMemberUpdate: user=%s roles=%v (no prior state)", m.User.ID, m.Roles)
		return
	}
	added, removed := diffRoles(m.BeforeUpdate.Roles, m.Roles)
	if len(added) == 0 && len(removed) == 0 {
		return
	}
	logger.SugarLogger.Infof("GuildMemberUpdate: user=%s added=%v removed=%v", m.User.ID, added, removed)
}

func OnGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	logger.SugarLogger.Infof("GuildMemberRemove: user=%s", m.User.ID)
}

func diffRoles(before, after []string) (added, removed []string) {
	beforeSet := make(map[string]struct{}, len(before))
	for _, r := range before {
		beforeSet[r] = struct{}{}
	}
	afterSet := make(map[string]struct{}, len(after))
	for _, r := range after {
		afterSet[r] = struct{}{}
	}
	for r := range afterSet {
		if _, ok := beforeSet[r]; !ok {
			added = append(added, r)
		}
	}
	for r := range beforeSet {
		if _, ok := afterSet[r]; !ok {
			removed = append(removed, r)
		}
	}
	return added, removed
}
