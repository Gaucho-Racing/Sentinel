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

	logger.SugarLogger.Infof("Message from %s in %s: %s", m.Author.ID, m.ChannelID, m.Content)

	_, err := service.CreateDiscordMessage(model.DiscordMessage{
		DiscordUserID: m.Author.ID,
		ChannelID:     m.ChannelID,
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
	default:
		logger.SugarLogger.Infof("Unknown command: %s", command)
	}
}

func OnDiscordReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	logger.SugarLogger.Infof("Reaction from %s in %s: %s", r.UserID, r.ChannelID, r.Emoji.Name)

	_, err := service.CreateDiscordReaction(model.DiscordReaction{
		DiscordUserID: r.UserID,
		ChannelID:     r.ChannelID,
		MessageID:     r.MessageID,
		Emoji:         r.Emoji.Name,
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to persist discord reaction: %v", err)
	}
}
