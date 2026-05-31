package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func InitializeBot() {
	if service.Discord == nil {
		logger.SugarLogger.Errorln("Discord session is not connected")
		return
	}
	service.Discord.AddHandler(OnDiscordMessage)
	service.Discord.AddHandler(OnDiscordReaction)
	service.Discord.AddHandler(OnGuildMemberAdd)
	service.Discord.AddHandler(OnGuildMemberUpdate)
	service.Discord.AddHandler(OnGuildMemberRemove)
	service.Discord.AddHandler(OnUserUpdate)
	service.Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err := service.Discord.Open()
	if err != nil {
		logger.SugarLogger.Errorln("Error opening Discord connection:", err)
		return
	}
	logger.SugarLogger.Infof("Discord Bot is now running! [Prefix = %s]", config.DiscordPrefix)
}

func OnDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	channelName := service.GetChannelName(m.ChannelID)

	logger.SugarLogger.Infof("Message from %s in #%s: %s", m.Author.ID, channelName, m.Content)

	_, err := service.CreateDiscordMessage(model.DiscordMessage{
		DiscordUserID: m.Author.ID,
		ChannelID:     m.ChannelID,
		ChannelName:   channelName,
		MessageID:     m.ID,
		Content:       m.Content,
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to persist discord message: %v", err)
	}

	if !strings.HasPrefix(m.Content, config.DiscordPrefix) {
		return
	}
	parts := strings.Fields(m.Content[len(config.DiscordPrefix):])
	if len(parts) == 0 {
		return
	}
	command := parts[0]
	args := parts[1:]
	switch command {
	case "ping":
		Ping(args, s, m)
	case "verify":
		Verify(args, s, m)
	default:
		logger.SugarLogger.Infof("Unknown command: %s", command)
	}
}

func OnDiscordReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	channelName := service.GetChannelName(r.ChannelID)

	logger.SugarLogger.Infof("Reaction from %s in #%s: %s", r.UserID, channelName, r.Emoji.Name)

	_, err := service.CreateDiscordReaction(model.DiscordReaction{
		DiscordUserID: r.UserID,
		ChannelID:     r.ChannelID,
		ChannelName:   channelName,
		MessageID:     r.MessageID,
		Emoji:         r.Emoji.Name,
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to persist discord reaction: %v", err)
	}
}

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
	if m.User == nil {
		return
	}
	if m.BeforeUpdate == nil {
		logger.SugarLogger.Infof("GuildMemberUpdate: user=%s roles=%v (no prior state)", m.User.ID, m.Roles)
	} else {
		added, removed := diffRoles(m.BeforeUpdate.Roles, m.Roles)
		if len(added) > 0 || len(removed) > 0 {
			logger.SugarLogger.Infof("GuildMemberUpdate: user=%s added=%v removed=%v", m.User.ID, added, removed)
		}
	}

	service.SyncDiscordUserAvatar(m.User.ID, m.User.AvatarURL("256"))
}

func OnGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	logger.SugarLogger.Infof("GuildMemberRemove: user=%s", m.User.ID)
}

func OnUserUpdate(s *discordgo.Session, u *discordgo.UserUpdate) {
	if u == nil || u.User == nil {
		return
	}
	logger.SugarLogger.Infof("UserUpdate: user=%s username=%s", u.ID, u.Username)
	service.SyncDiscordUserAvatar(u.ID, u.AvatarURL("256"))
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
