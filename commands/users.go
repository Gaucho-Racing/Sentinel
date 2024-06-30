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
	leadMembers := 0
	officerMembers := 0

	subteamMap := make(map[string]int)
	subteams := service.GetAllSubteams()
	for _, subteam := range subteams {
		subteamMap[subteam.Name] = 0
	}

	for _, member := range members {
		user := service.GetUserByID(member.User.ID)
		if user.ID != "" {
			verifiedMembers++
		}
		for _, role := range member.Roles {
			if role == config.MemberRoleID {
				memberMembers++
			}
			if role == config.LeadRoleID {
				leadMembers++
			}
			if role == config.OfficerRoleID {
				officerMembers++
			}
			for _, subteam := range subteams {
				if role == subteam.ID {
					subteamMap[subteam.Name]++
				}
			}
		}
		guildMembers++
	}
	messageText := fmt.Sprintf("Total Members: %d\nMembers Role: %d\nVerified Members: %d\n\n", guildMembers, memberMembers, verifiedMembers)
	utils.SugarLogger.Infof("Total Members: %d", guildMembers)
	utils.SugarLogger.Infof("Members Role: %d", memberMembers)
	utils.SugarLogger.Infof("Verified Members: %d", verifiedMembers)
	for subteam, count := range subteamMap {
		utils.SugarLogger.Infof("%s: %d", subteam, count)
		messageText += fmt.Sprintf("%s: %d\n", subteam, count)
	}
	utils.SugarLogger.Infof("Lead Members: %d", leadMembers)
	utils.SugarLogger.Infof("Officer Members: %d", officerMembers)
	messageText += fmt.Sprintf("\nLeads: %d\nOfficers: %d", leadMembers, officerMembers)

	go service.Discord.ChannelMessageEdit(m.ChannelID, msg.ID, messageText)
}
