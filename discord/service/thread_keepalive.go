package service

import (
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

// Discord auto-archive durations, in minutes. Changing auto_archive_duration
// counts as thread activity, which is how we bump without posting a message.
const (
	maxAutoArchiveMinutes = 10080 // 7 days
	midAutoArchiveMinutes = 4320  // 3 days
)

// threadBumpInterval is how often we scan active threads. threadBumpLeadTime
// is how close to Discord's channel-list hide deadline we bump.
const (
	threadBumpInterval = 5 * time.Minute
	threadBumpLeadTime = 30 * time.Minute
)

var threadSweepRunning atomic.Bool

// KeepThreadAlive unarchives a thread that Discord auto-archived. Archived
// threads are immutable: Discord rejects any PATCH other than archived=false
// (error 50083), so the 7-day window is set in a second edit after unarchive
// succeeds. Requires MANAGE_THREADS in the guild.
func KeepThreadAlive(thread *discordgo.Channel) {
	if thread == nil {
		return
	}
	archived := false
	if _, err := Discord.ChannelEdit(thread.ID, &discordgo.ChannelEdit{
		Archived: &archived,
	}); err != nil {
		logger.SugarLogger.Errorf("thread keepalive: failed to unarchive thread %s (%s): %v", thread.ID, thread.Name, err)
		return
	}
	logger.SugarLogger.Infof("thread keepalive: unarchived thread %s (%s)", thread.ID, thread.Name)

	if thread.ThreadMetadata == nil || thread.ThreadMetadata.AutoArchiveDuration != maxAutoArchiveMinutes {
		if _, err := Discord.ChannelEdit(thread.ID, &discordgo.ChannelEdit{
			AutoArchiveDuration: maxAutoArchiveMinutes,
		}); err != nil {
			logger.SugarLogger.Errorf("thread keepalive: failed to extend auto-archive on thread %s (%s): %v", thread.ID, thread.Name, err)
		}
	}
}

// StartThreadKeepaliveCron spawns a background goroutine that periodically
// bumps active threads approaching Discord's channel-list hide deadline.
// Changing auto_archive_duration counts as activity, so the bump does not
// post a message. ThreadUpdate unarchives any that still slip through.
//
// A first sweep runs immediately so a restart recovers threads that went
// stale while we were down. If a sweep is still running when the next tick
// fires, the tick is skipped. The goroutine is leaked on process exit; we
// don't have graceful-shutdown plumbing for it and the OS reaps everything
// anyway.
func StartThreadKeepaliveCron() {
	logger.SugarLogger.Infof("thread keepalive: cron enabled, interval=%v", threadBumpInterval)
	go func() {
		sweepStaleThreads()
		ticker := time.NewTicker(threadBumpInterval)
		defer ticker.Stop()
		for range ticker.C {
			sweepStaleThreads()
		}
	}()
}

func sweepStaleThreads() {
	if Discord == nil {
		return
	}
	if !threadSweepRunning.CompareAndSwap(false, true) {
		return
	}
	defer threadSweepRunning.Store(false)

	list, err := Discord.GuildThreadsActive(config.DiscordGuild)
	if err != nil {
		logger.SugarLogger.Errorf("thread keepalive: failed to list active threads: %v", err)
		return
	}
	if list == nil {
		return
	}

	now := time.Now()
	bumped := 0
	for _, thread := range list.Threads {
		if thread == nil || thread.ThreadMetadata == nil || thread.ThreadMetadata.Locked {
			continue
		}
		if thread.ThreadMetadata.Archived {
			KeepThreadAlive(thread)
			continue
		}
		if threadIsStale(thread.ThreadMetadata, now) && bumpThread(thread) {
			bumped++
		}
	}
	if bumped > 0 {
		logger.SugarLogger.Infof("thread keepalive: bumped %d stale thread(s)", bumped)
	}
}

// threadIsStale reports whether the thread will leave the channel list within
// threadBumpLeadTime, based on Discord's last-activity timestamp.
func threadIsStale(meta *discordgo.ThreadMetadata, now time.Time) bool {
	if meta == nil {
		return false
	}
	dur := time.Duration(meta.AutoArchiveDuration) * time.Minute
	if dur <= 0 {
		dur = time.Hour
	}
	return !now.Before(meta.ArchiveTimestamp.Add(dur - threadBumpLeadTime))
}

// bumpThread toggles auto_archive_duration. Discord treats that change as
// activity, which resets the hide timer without posting a message.
func bumpThread(thread *discordgo.Channel) bool {
	if thread == nil {
		return false
	}
	next := midAutoArchiveMinutes
	if thread.ThreadMetadata != nil && thread.ThreadMetadata.AutoArchiveDuration != maxAutoArchiveMinutes {
		next = maxAutoArchiveMinutes
	}
	if _, err := Discord.ChannelEdit(thread.ID, &discordgo.ChannelEdit{
		AutoArchiveDuration: next,
	}); err != nil {
		logger.SugarLogger.Errorf("thread keepalive: failed to bump thread %s (%s): %v", thread.ID, thread.Name, err)
		return false
	}
	logger.SugarLogger.Debugf("thread keepalive: bumped thread %s (%s) auto_archive=%d", thread.ID, thread.Name, next)
	return true
}
