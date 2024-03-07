package commands

import (
	"github.com/bwmarrin/discordgo"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
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
	println("Message received: " + m.Content)
}
