package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

const unarchiveReplyTTL = 10 * time.Second

func Unarchive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	permissions, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		logger.SugarLogger.Errorf("unarchive: failed to compute permissions for %s in %s: %v", m.Author.ID, m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> something went wrong, try again in a minute.", m.Author.ID), unarchiveReplyTTL)
		return
	}
	if permissions&discordgo.PermissionManageChannels == 0 {
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you need the Manage Channels permission to unarchive this channel.", m.Author.ID), unarchiveReplyTTL)
		return
	}

	record, err := service.UnarchiveChannel(m.ChannelID)
	if err != nil {
		logger.SugarLogger.Errorf("unarchive: failed for channel %s: %v", m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> this channel isn't archived, or restoring it failed — check the logs.", m.Author.ID), unarchiveReplyTTL)
		return
	}

	content := "This channel has been unarchived and its permissions restored."
	if record.PreviousParentID == "" {
		content += " I couldn't determine its original category, so it'll need to be moved manually."
	}
	if _, err := s.ChannelMessageSend(m.ChannelID, content); err != nil {
		logger.SugarLogger.Errorf("unarchive: failed to send confirmation in %s: %v", m.ChannelID, err)
	}
}
