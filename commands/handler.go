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
	service.Discord.AddHandler(OnGuildMemberUpdate)
	service.Discord.AddHandler(LogUserMessage)
	service.Discord.AddHandler(LogUserReaction)
	service.Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
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
	case "rs":
		RemoveSubteam(args, s, m)
	case "github":
		Github(args, s, m)
	case "drive":
		Drive(args, s, m)
	case "whois":
		Whois(args, s, m)
	case "users":
		Users(args, s, m)
	default:
		utils.SugarLogger.Infof("Command not found: %s", command)
	}
}

func OnGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	utils.SugarLogger.Infof("Member update: (%s) %s", m.User.ID, m.Nick)
	newRoles := m.Roles
	user := service.GetUserByID(m.User.ID)
	if user.ID == "" {
		return
	}
	service.SyncDiscordRolesForUser(user.ID, newRoles)
}

func LogUserMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	utils.SugarLogger.Infof("Message from %s in %s: %s", m.Author.ID, m.ChannelID, m.Content)
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

func LogUserReaction(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	utils.SugarLogger.Infof("Reaction from %s in %s: %s", m.UserID, m.ChannelID, m.Emoji.Name)
	// Get user info
	user := service.GetUserByID(m.UserID)
	if user.ID == "" {
		return
	}
	// Log reaction
	service.CreateActivity(model.UserActivity{
		ID:     uuid.New().String(),
		UserID: user.ID,
		Action: "reaction",
	})
}
