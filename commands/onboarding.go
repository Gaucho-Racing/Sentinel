package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

var onboardingMechRole = "1293019097650696323"
var onboardingElecRole = "1293019255159394428"

func Onboarding(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	if m.GuildID != config.DiscordGuild {
		m.GuildID = config.DiscordGuild
	}
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	isOfficer := false
	for _, role := range guildMember.Roles {
		if role == "812948550819905546" {
			isOfficer = true
			break
		}
	}
	utils.SugarLogger.Infof("User %s is officer: %t", m.Author.ID, isOfficer)

	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
		return
	} else {
		if len(args) < 1 {
			go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!onboarding <mech | elec>`", 5*time.Second)
			return
		}
		switch args[0] {
		case "mech":
			err = s.GuildMemberRoleAdd(m.GuildID, user.ID, onboardingMechRole)
			if err != nil {
				utils.SugarLogger.Errorln(err)
				go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
				return
			}
		case "elec":
			err = s.GuildMemberRoleAdd(m.GuildID, user.ID, onboardingElecRole)
			if err != nil {
				utils.SugarLogger.Errorln(err)
				go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
				return
			}
		default:
			go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!onboarding <mech | elec>`", 5*time.Second)
			return
		}
	}
}
