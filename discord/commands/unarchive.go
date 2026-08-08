package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

const archiveReplyTTL = 10 * time.Second

// requireArchiveAccess gates the archive commands to members of the allowed
// Sentinel groups, replying with a disappearing message when the check
// fails. Fails closed: a missing entity link or a core lookup failure both
// deny access.
func requireArchiveAccess(s *discordgo.Session, m *discordgo.MessageCreate, command string) bool {
	groupNames, err := service.GetGroupNamesForDiscordUser(m.Author.ID)
	if err != nil {
		logger.SugarLogger.Errorf("%s: failed to fetch sentinel groups for %s: %v", command, m.Author.ID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you don't have permission to %s channels.", m.Author.ID, command), archiveReplyTTL)
		return false
	}
	for _, name := range groupNames {
		for _, allowed := range config.ArchiveCommandAllowedGroups {
			if strings.EqualFold(name, allowed) {
				return true
			}
		}
	}
	service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you don't have permission to %s channels.", m.Author.ID, command), archiveReplyTTL)
	return false
}

func Unarchive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !requireArchiveAccess(s, m, "unarchive") {
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
