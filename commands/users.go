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
	user := service.GetUserByID(guildMember.User.ID)
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
	leadMembers := 0
	officerMembers := 0
	specialAdvisorMembers := 0
	alumniMembers := 0
	teamMembers := 0

	subteamMap := make(map[string]int)
	subteams := service.GetAllSubteams()
	for _, subteam := range subteams {
		subteamMap[subteam.Name] = 0
	}

	for _, member := range members {
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
			if role == config.SpecialAdvisorRoleID {
				specialAdvisorMembers++
			}
			if role == config.AlumniRoleID {
				alumniMembers++
			}
			if role == config.TeamMemberRoleID {
				teamMembers++
			}
			for _, subteam := range subteams {
				if role == subteam.ID {
					subteamMap[subteam.Name]++
				}
			}
		}
		guildMembers++
	}
	messageText := fmt.Sprintf("Discord Members: %d\nMembers Role: %d\nAlumni Members: %d\nTeam Members: %d\n\n", guildMembers, memberMembers, alumniMembers, teamMembers)
	utils.SugarLogger.Infof("Discord Members: %d", guildMembers)
	utils.SugarLogger.Infof("Members Role: %d", memberMembers)
	utils.SugarLogger.Infof("Alumni Members: %d", alumniMembers)
	utils.SugarLogger.Infof("Team Members: %d", teamMembers)
	for subteam, count := range subteamMap {
		utils.SugarLogger.Infof("%s: %d", subteam, count)
		messageText += fmt.Sprintf("%s: %d\n", subteam, count)
	}
	utils.SugarLogger.Infof("Lead Members: %d", leadMembers)
	utils.SugarLogger.Infof("Officer Members: %d", officerMembers)
	utils.SugarLogger.Infof("Special Advisor Members: %d", specialAdvisorMembers)
	messageText += fmt.Sprintf("\nLeads: %d\nOfficers: %d\nSpecial Advisors: %d", leadMembers, officerMembers, specialAdvisorMembers)

	go service.Discord.ChannelMessageEdit(m.ChannelID, msg.ID, messageText)
}
