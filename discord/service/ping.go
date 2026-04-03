package service

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
)

func Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	message, err := s.ChannelMessageSend(m.ChannelID, "Pong from "+config.Service.FormattedNameWithVersion()+"!")
	if err == nil {
		delay := message.Timestamp.Sub(m.Timestamp).Milliseconds()
		s.ChannelMessageEdit(m.ChannelID, message.ID, "Pong from "+config.Service.FormattedNameWithVersion()+"! (**"+strconv.FormatInt(delay, 10)+"ms**)")
	}
}
