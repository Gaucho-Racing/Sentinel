package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

const archiveReplyTTL = 10 * time.Second

// requireManageChannels gates the archive commands behind the Manage
// Channels permission in the invoking channel, replying with a disappearing
// message when the check fails.
func requireManageChannels(s *discordgo.Session, m *discordgo.MessageCreate, command string) bool {
	permissions, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		logger.SugarLogger.Errorf("%s: failed to compute permissions for %s in %s: %v", command, m.Author.ID, m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> something went wrong, try again in a minute.", m.Author.ID), archiveReplyTTL)
		return false
	}
	if permissions&discordgo.PermissionManageChannels == 0 {
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you need the Manage Channels permission to %s this channel.", m.Author.ID, command), archiveReplyTTL)
		return false
	}
	return true
}

func Unarchive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !requireManageChannels(s, m, "unarchive") {
		return
	}

	record, err := service.UnarchiveChannel(m.ChannelID)
	if err != nil {
		logger.SugarLogger.Errorf("unarchive: failed for channel %s: %v", m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> this channel isn't archived, or restoring it failed — check the logs.", m.Author.ID), archiveReplyTTL)
		return
	}

	content := "This channel has been unarchived and its permissions restored."
	if record.PreviousParentID == "" {
		content += " It wasn't in a category before it was archived, so it'll need to be moved out manually."
	}
	if _, err := s.ChannelMessageSend(m.ChannelID, content); err != nil {
		logger.SugarLogger.Errorf("unarchive: failed to send confirmation in %s: %v", m.ChannelID, err)
	}
}
