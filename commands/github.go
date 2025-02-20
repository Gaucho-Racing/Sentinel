package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Github(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
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
		return
	}
	if len(args) < 1 {
		go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!github <username>`", 5*time.Second)
		return
	}
	err = service.AddUserToGithub(m.Author.ID, args[0])
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	go service.SendDisappearingMessage(m.ChannelID, "Successfully invited user to GitHub organization!", 5*time.Second)
}
