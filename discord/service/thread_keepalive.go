package service

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

const threadReopenNotice = "This thread was archived due to inactivity, so I've automatically reopened it. " +
	"If this thread is no longer needed, lock it before closing and I'll leave it alone."

// KeepThreadAlive unarchives a thread that Discord just auto-archived and
// bumps its auto-archive window to the maximum (7 days) so the gateway only
// re-archives it weekly instead of on the channel's default window. A notice
// explaining the reopen and the lock-to-close opt-out is posted in the thread
// and auto-deleted after an hour. Requires the bot to have MANAGE_THREADS in
// the guild.
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
	SendDisappearingMessage(thread.ID, threadReopenNotice, time.Hour)
}
