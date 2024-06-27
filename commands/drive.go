package commands

import (
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Drive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}

	user := service.GetUserByID(guildMember.User.ID)
	if user.ID == "" {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
	} else {
		service.GetDriveMemberStatus(user.Email)
	}
}
