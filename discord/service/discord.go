package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

var Discord *discordgo.Session

func ConnectDiscord() {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		logger.SugarLogger.Errorln("Error creating Discord session:", err)
		return
	}
	Discord = dg
	logger.SugarLogger.Infoln("Created Discord session")
}

func InitializeBot() {
	Discord.AddHandler(OnDiscordMessage)
	Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err := Discord.Open()
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
	if len(m.Content) < len(config.DiscordPrefix) || m.Content[:len(config.DiscordPrefix)] != config.DiscordPrefix {
		return
	}
	args := splitCommand(m.Content[len(config.DiscordPrefix):])
	if len(args) == 0 {
		return
	}
	command := args[0]
	args = args[1:]
	switch command {
	case "ping":
		Ping(s, m)
	default:
		logger.SugarLogger.Infof("Unknown command: %s", command)
	}
}

func splitCommand(s string) []string {
	var result []string
	for _, part := range splitBySpace(s) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitBySpace(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
