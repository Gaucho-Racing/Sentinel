package commands

import (
	"fmt"
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"slices"
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
	defer ChannelMessageFilter(s, m)
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
	case "alumni":
		Alumni(args, s, m)
	default:
		utils.SugarLogger.Infof("Command not found: %s", command)
	}
}

func OnGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID != config.DiscordGuild {
		utils.SugarLogger.Infof("Recieved member update event for guild %s, ignoring...", m.GuildID)
		service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Recieved member update event for guild %s, ignoring...", m.GuildID))
		return
	}
	if m.User.Bot {
		utils.SugarLogger.Infof("Recieved member update event for bot %s (%s), ignoring...", m.User.ID, m.Nick)
		service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Recieved member update event for bot %s (%s), ignoring...", m.User.ID, m.Nick))
		return
	}
	utils.SugarLogger.Infof("Member update: (%s) %s", m.User.ID, m.Nick)
	service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Member update: (%s) %s", m.User.ID, m.Nick))
	newRoles := m.Roles
	user := service.GetUserByID(m.User.ID)
	if user.ID == "" {
		// User is not in Sentinel, ensure they cannot have any roles
		service.SetDiscordRolesForUser(m.User.ID, []string{})
		return
	}

	// Verify discord specific role rules
	// If user is alumni, they cannot have any subteam roles
	if slices.Contains(newRoles, config.AlumniRoleID) {
		// Remove all subteam roles
		for _, role := range service.GetAllSubteams() {
			err := service.Discord.GuildMemberRoleRemove(config.DiscordGuild, m.User.ID, role.ID)
			if err != nil {
				utils.SugarLogger.Errorf("Error removing subteam role %s from user %s (%s): %s", role.ID, m.User.ID, m.Nick, err)
				service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Error removing subteam role %s from user %s (%s): %s", role.ID, m.User.ID, m.Nick, err))
			}
		}
		utils.SugarLogger.Infof("Removed all subteam roles from user %s (%s) as they are alumni", m.User.ID, m.Nick)
		service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removed all subteam roles from user %s (%s) as they are alumni", m.User.ID, m.Nick))

		// User cannot have member, lead, or officer roles if they are alumni (admin and special advisor ok)
		removeRoles := []string{config.MemberRoleID, config.LeadRoleID, config.OfficerRoleID}
		for _, role := range removeRoles {
			err := service.Discord.GuildMemberRoleRemove(config.DiscordGuild, m.User.ID, role)
			if err != nil {
				utils.SugarLogger.Errorf("Error removing role %s from user %s (%s): %s", role, m.User.ID, m.Nick, err)
				service.SendMessage(config.DiscordLogChannel, fmt.Sprintf("Error removing role %s from user %s (%s): %s", role, m.User.ID, m.Nick, err))
			}
		}
	}

	// If user is not alumni or member, remove all roles
	if !slices.Contains(newRoles, config.AlumniRoleID) && !slices.Contains(newRoles, config.MemberRoleID) {
		service.SetDiscordRolesForUser(m.User.ID, []string{})
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

func ChannelMessageFilter(s *discordgo.Session, m *discordgo.MessageCreate) {
	var verificationChannel = "1215484329736671282"
	var rolesChannel = "1215525696286232626"

	channels := []string{verificationChannel, rolesChannel}

	for _, channel := range channels {
		if m.ChannelID == channel {
			utils.SugarLogger.Infof("Deleting message from %s in %s: %s", m.Author.ID, m.ChannelID, m.Content)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
	}
}
