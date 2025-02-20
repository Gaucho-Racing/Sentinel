package commands

import (
	"fmt"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Drive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	user := service.GetUserByID(guildMember.User.ID)
	if user.ID == "" || !user.IsMember() || !user.IsAlumni() {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
	} else {
		loadingMessage, _ := s.ChannelMessageSend(m.ChannelID, "checking drive access...")
		role := "writer"
		if user.IsInnerCircle() {
			role = "organizer"
		}
		perm, _ := service.GetDriveMemberPermission(config.SharedDriveID, user.Email)
		if perm != nil {
			// Remove and re-add user to update role
			_ = service.RemoveMemberFromDrive(config.SharedDriveID, user.Email)
			_ = service.AddMemberToDrive(config.SharedDriveID, user.Email, role)
			perm, _ := service.GetDriveMemberPermission(config.SharedDriveID, user.Email)
			service.Discord.ChannelMessageDelete(m.ChannelID, loadingMessage.ID)
			go service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("You already have `%s` access to the shared drive!", perm.Role), 5*time.Second)
		} else {
			err = service.AddMemberToDrive(config.SharedDriveID, user.Email, role)
			if err != nil {
				utils.SugarLogger.Errorln(err)
				service.Discord.ChannelMessageDelete(m.ChannelID, loadingMessage.ID)
				go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
			} else {
				service.Discord.ChannelMessageDelete(m.ChannelID, loadingMessage.ID)
				go service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("You have been added to the shared drive with `%s` access!", role), 5*time.Second)
			}
		}
	}
}
