package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Say(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	if m.GuildID != config.DiscordGuild {
		m.GuildID = config.DiscordGuild
	}
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	if !utils.IsInnerCircle(guildMember.Roles) {
		go service.SendDisappearingMessage(m.ChannelID, "You do not have access to this command!", 5*time.Second)
		return
	}
	if len(args) < 1 {
		if len(m.Attachments) > 0 {
			SendAttachments(m.Attachments, s, m)
			return
		}
		go service.SendDisappearingMessage(m.ChannelID, "Must be in the format: `!say <message>`", 5*time.Second)
		return
	}
	message, _ := strings.CutPrefix(m.Content, config.Prefix+"say ")
	s.ChannelMessageSend(m.ChannelID, message)
	SendAttachments(m.Attachments, s, m)
}

func SendAttachments(attachments []*discordgo.MessageAttachment, s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, attachment := range attachments {
		println(attachment.URL)
		s.ChannelMessageSend(m.ChannelID, attachment.URL)
	}
}
