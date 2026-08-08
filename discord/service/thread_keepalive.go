package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

// KeepThreadAlive unarchives a thread that Discord just auto-archived and
// bumps its auto-archive window to the maximum (7 days) so the gateway only
// re-archives it weekly instead of on the channel's default window. The
// unarchive is silent — no message is posted and members aren't notified.
// Requires the bot to have MANAGE_THREADS in the guild.
func KeepThreadAlive(thread *discordgo.Channel) {
	archived := false
	_, err := Discord.ChannelEdit(thread.ID, &discordgo.ChannelEdit{
		Archived:            &archived,
		AutoArchiveDuration: 10080,
	})
	if err != nil {
		logger.SugarLogger.Errorf("thread keepalive: failed to unarchive thread %s (%s): %v", thread.ID, thread.Name, err)
		return
	}
	logger.SugarLogger.Infof("thread keepalive: unarchived thread %s (%s)", thread.ID, thread.Name)
}
