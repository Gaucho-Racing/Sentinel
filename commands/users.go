package commands

import (
	"fmt"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Users(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	utils.SugarLogger.Infof("User: %s", guildMember.User.ID)
	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
		return
	}
	msg, _ := service.Discord.ChannelMessageSend(m.ChannelID, "Fetching user data...")
	members, err := s.GuildMembers(m.GuildID, "", 1000)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	guildMembers := 0
	memberMembers := 0
	verifiedMembers := 0
	aeroMembers := 0
	controlsMembers := 0
	chassisMembers := 0
	suspensionMembers := 0
	powertrainMembers := 0
	businessMembers := 0
	leadMembers := 0
	officerMembers := 0

	for _, member := range members {
		user := service.GetUserByID(member.User.ID)
		if user.ID != "" {
			verifiedMembers++
		}
		for _, role := range member.Roles {
			if role == config.MemberRoleID {
				memberMembers++
			}
			if role == service.GetSubteamByName("Aero").ID {
				aeroMembers++
			}
			if role == service.GetSubteamByName("Controls").ID {
				controlsMembers++
			}
			if role == service.GetSubteamByName("Chassis").ID {
				chassisMembers++
			}
			if role == service.GetSubteamByName("Suspension").ID {
				suspensionMembers++
			}
			if role == service.GetSubteamByName("Powertrain").ID {
				powertrainMembers++
			}
			if role == service.GetSubteamByName("Business").ID {
				businessMembers++
			}
			if role == config.LeadRoleID {
				leadMembers++
			}
			if role == config.OfficerRoleID {
				officerMembers++
			}
		}
		guildMembers++
	}
	utils.SugarLogger.Infof("Total Members: %d", guildMembers)
	utils.SugarLogger.Infof("Members Role: %d", memberMembers)
	utils.SugarLogger.Infof("Verified Members: %d", verifiedMembers)
	utils.SugarLogger.Infof("Aero Members: %d", aeroMembers)
	utils.SugarLogger.Infof("Business Members: %d", businessMembers)
	utils.SugarLogger.Infof("Chassis Members: %d", chassisMembers)
	utils.SugarLogger.Infof("Controls Members: %d", controlsMembers)
	utils.SugarLogger.Infof("Suspension Members: %d", suspensionMembers)
	utils.SugarLogger.Infof("Powertrain Members: %d", powertrainMembers)
	utils.SugarLogger.Infof("Lead Members: %d", leadMembers)
	utils.SugarLogger.Infof("Officer Members: %d", officerMembers)

	go service.Discord.ChannelMessageEdit(m.ChannelID, msg.ID, fmt.Sprintf("Total Members: %d\nMembers Role: %d\nVerified Members: %d\n\nAero: %d\nBusiness: %d\nChassis: %d\nControls: %d\nSuspension: %d\nPowertrain: %d\n\nLeads: %d\nOfficers: %d", guildMembers, memberMembers, verifiedMembers, aeroMembers, businessMembers, chassisMembers, controlsMembers, suspensionMembers, powertrainMembers, leadMembers, officerMembers))
}
