package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func Archive(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !requireArchiveAccess(s, m, "archive") {
		return
	}
	if _, err := service.GetArchivedChannel(m.ChannelID); err == nil {
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> this channel is already archived.", m.Author.ID), archiveReplyTTL)
		return
	}
	if err := service.ArchiveChannel(m.ChannelID, m.Author.ID); err != nil {
		logger.SugarLogger.Errorf("archive: failed for channel %s: %v", m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> archiving failed — check the logs.", m.Author.ID), archiveReplyTTL)
	}
}
