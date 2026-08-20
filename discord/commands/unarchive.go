package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func Unarchive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	allowedGroups := []string{"Admins", "Leads", "Officers"}
	if !requireGroupMembership(m, "unarchive", allowedGroups) {
		return
	}
	if !requireNotThread(s, m, "unarchive") {
		return
	}

	record, err := service.UnarchiveChannel(m.ChannelID)
	if err != nil {
		logger.SugarLogger.Errorf("unarchive: failed for channel %s: %v", m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> this channel isn't archived, or restoring it failed — check the logs.", m.Author.ID), commandReplyTTL)
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
