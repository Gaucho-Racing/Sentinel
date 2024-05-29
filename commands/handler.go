package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func InitializeDiscordBot() {
	service.Discord.AddHandler(OnDiscordMessage)
	service.Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	err := service.Discord.Open()
	if err != nil {
		utils.SugarLogger.Errorln("error opening connection,", err)
		return
	}
	utils.SugarLogger.Infoln("Discord Bot is now running! [Prefix = " + config.Prefix + "]")
}

func OnDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// or messages that don't start with the prefix
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, config.Prefix) {
		return
	}
	command := strings.Split(m.Content, " ")[0][len(config.Prefix):]
	args := strings.Split(m.Content, " ")[1:]
	switch command {
	case "ping":
		Ping(args, s, m)
	case "say":
		Say(args, s, m)
	case "verify":
		Verify(args, s, m)
	case "subteam":
		Subteam(args, s, m)
	default:
		utils.SugarLogger.Infoln("Command not found: %s", command)
	}
}
