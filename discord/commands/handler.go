package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func InitializeBot() {
	if service.Discord == nil {
		logger.SugarLogger.Errorln("Discord session is not connected")
		return
	}
	service.Discord.AddHandler(OnDiscordMessage)
	service.Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err := service.Discord.Open()
	if err != nil {
		logger.SugarLogger.Errorln("Error opening Discord connection:", err)
		return
	}
	logger.SugarLogger.Infof("Discord Bot is now running! [Prefix = %s]", config.DiscordPrefix)
}

func OnDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
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
