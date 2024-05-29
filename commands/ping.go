package commands

import (
	"sentinel/config"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func Ping(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	message, err := s.ChannelMessageSend(m.ChannelID, "Pong from Sentinel v"+config.Version+"!")
	if err == nil {
		delay := message.Timestamp.Sub(m.Timestamp).Milliseconds()
		s.ChannelMessageEdit(m.ChannelID, message.ID, "Pong from Sentinel v"+config.Version+"!  (**"+strconv.FormatInt(delay, 10)+"ms**)")
	}
}
