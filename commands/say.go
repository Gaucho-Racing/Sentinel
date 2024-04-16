package commands

import (
	"github.com/bwmarrin/discordgo"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"
)

func Say(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	isOfficer := false
	for _, role := range guildMember.Roles {
		if role == "812948550819905546" || role == "970423652791246888" {
			isOfficer = true
			break
		}
	}
	if !isOfficer {
		go service.SendDisappearingMessage(m.ChannelID, "You must be an officer or team lead to use this command!", 5*time.Second)
		return
	}
	if len(args) < 1 {
		go service.SendDisappearingMessage(m.ChannelID, "Must be in the format: `!say <message>`", 5*time.Second)
		return
	}
	message, _ := strings.CutPrefix(m.Content, config.Prefix+"say ")
	s.ChannelMessageSend(m.ChannelID, message)
}
