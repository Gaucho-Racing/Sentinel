package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

// Say reposts a message as the bot, so announcements come from Sentinel
// rather than whoever typed them. The command message is always deleted,
// including when access is denied, so a failed attempt doesn't linger.
func Say(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	if !requireGroupMembership(m, "say", []string{"Admins", "DevopsMembers"}) {
		return
	}
	// Cut from the raw content rather than joining args, so line breaks and
	// spacing in the announcement survive.
	message := strings.TrimSpace(strings.TrimPrefix(m.Content, config.DiscordPrefix+"say"))
	if message == "" {
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> usage: `%ssay <message>`", m.Author.ID, config.DiscordPrefix), commandReplyTTL)
		return
	}
	if _, err := s.ChannelMessageSend(m.ChannelID, message); err != nil {
		logger.SugarLogger.Errorf("say: failed to send in %s: %v", m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> couldn't send that message — check the logs.", m.Author.ID), commandReplyTTL)
	}
}
