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
	case "github":
		Github(args, s, m)
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
	subteamRoles := make([]model.UserSubteam, 0)
	roles := service.GetRolesForUser(m.User.ID)
	for i, role := range roles {
		if strings.HasPrefix(role, "d_") {
			roles = append(roles[:i], roles[i+1:]...)
		}
	}
	for _, id := range newRoles {
		subteam := service.GetSubteamByID(id)
		if subteam.ID != "" {
			subteamRoles = append(subteamRoles, model.UserSubteam{
				UserID: user.ID,
				RoleID: subteam.ID,
			})
		} else if id == config.AdminRoleID {
			roles = append(roles, "d_admin")
		} else if id == config.OfficerRoleID {
			roles = append(roles, "d_officer")
		} else if id == config.LeadRoleID {
			roles = append(roles, "d_lead")
		} else if id == config.MemberRoleID {
			roles = append(roles, "d_member")
		}
	}
	service.SetSubteamsForUser(user.ID, subteamRoles)
	service.SetRolesForUser(user.ID, roles)
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
