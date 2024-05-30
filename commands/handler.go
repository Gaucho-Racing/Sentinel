package commands

import (
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
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
	LogUserMessage(s, m)
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
	case "github":
		Github(args, s, m)
	case "whois":
		Whois(args, s, m)
	default:
		utils.SugarLogger.Infof("Command not found: %s", command)
	}
}

func LogUserMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	utils.SugarLogger.Infof("Message from %s: %s", m.Author.ID, m.Content)
	// Get user info
	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" {
		return
	}
	// Log message
	service.CreateActivity(model.UserActivity{
		ID:     uuid.New().String(),
		UserID: user.ID,
		Action: "message",
	})
}
