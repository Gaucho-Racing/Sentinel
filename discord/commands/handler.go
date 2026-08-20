package commands

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

const commandReplyTTL = 10 * time.Second

// requireGroupMembership gates a command to members of the given Sentinel
// groups (matched by name, case-insensitive), replying with a disappearing
// message when the check fails. Fails closed: a missing entity link or a
// core lookup failure both deny access.
func requireGroupMembership(m *discordgo.MessageCreate, command string, allowedGroups []string) bool {
	groupNames, err := service.GetGroupNamesForDiscordUser(m.Author.ID)
	if err != nil {
		logger.SugarLogger.Errorf("%s: failed to fetch sentinel groups for %s: %v", command, m.Author.ID, err)
	} else {
		for _, name := range groupNames {
			for _, allowed := range allowedGroups {
				if strings.EqualFold(name, allowed) {
					return true
				}
			}
		}
	}
	service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you don't have permission to use the `%s%s` command.", m.Author.ID, config.DiscordPrefix, command), commandReplyTTL)
	return false
}

// requireNotThread rejects commands run in a thread, forum post, or forum/media channel.
func requireNotThread(s *discordgo.Session, m *discordgo.MessageCreate, command string) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
	}
	if err != nil {
		logger.SugarLogger.Errorf("%s: failed to fetch channel %s: %v", command, m.ChannelID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> something went wrong, try again in a minute.", m.Author.ID), commandReplyTTL)
		return false
	}
	switch channel.Type {
	case discordgo.ChannelTypeGuildNewsThread, discordgo.ChannelTypeGuildPublicThread, discordgo.ChannelTypeGuildPrivateThread, discordgo.ChannelTypeGuildForum, discordgo.ChannelTypeGuildMedia:
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> `%s%s` can't be used in a thread.", m.Author.ID, config.DiscordPrefix, command), commandReplyTTL)
		return false
	}
	return true
}

// readyOnce guards the startup sweep so a gateway reconnect (which also
// fires Ready) doesn't repeatedly kick the sweep. Subsequent reconnects
// are covered by the periodic cron + per-user event reconciles anyway.
var readyOnce sync.Once

func InitializeBot() {
	if service.Discord == nil {
		logger.SugarLogger.Errorln("Discord session is not connected")
		return
	}
	service.Discord.AddHandler(OnReady)
	service.Discord.AddHandler(OnDiscordMessage)
	service.Discord.AddHandler(OnDiscordReaction)
	service.Discord.AddHandler(OnGuildMemberAdd)
	service.Discord.AddHandler(OnGuildMemberUpdate)
	service.Discord.AddHandler(OnGuildMemberRemove)
	service.Discord.AddHandler(OnUserUpdate)
	service.Discord.AddHandler(OnThreadUpdate)
	service.Discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err := service.Discord.Open()
	if err != nil {
		logger.SugarLogger.Errorln("Error opening Discord connection:", err)
		return
	}
	logger.SugarLogger.Infof("Discord Bot is now running! [Prefix = %s]", config.DiscordPrefix)
}

func OnDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	channelName := service.GetChannelName(m.ChannelID)

	logger.SugarLogger.Infof("Message from %s in #%s: %s", m.Author.ID, channelName, m.Content)

	_, err := service.CreateDiscordMessage(model.DiscordMessage{
		DiscordUserID: m.Author.ID,
		ChannelID:     m.ChannelID,
		ChannelName:   channelName,
		MessageID:     m.ID,
		Content:       m.Content,
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to persist discord message: %v", err)
	}

	if enforceVerificationChannel(s, m) {
		return
	}

	if !strings.HasPrefix(m.Content, config.DiscordPrefix) {
		return
	}
	parts := strings.Fields(m.Content[len(config.DiscordPrefix):])
	if len(parts) == 0 {
		return
	}
	command := parts[0]
	args := parts[1:]
	switch command {
	case "ping":
		Ping(args, s, m)
	case "verify":
		Verify(args, s, m)
	case "archive":
		Archive(args, s, m)
	case "unarchive":
		Unarchive(args, s, m)
	default:
		logger.SugarLogger.Infof("Unknown command: %s", command)
	}
}

func OnDiscordReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	channelName := service.GetChannelName(r.ChannelID)

	logger.SugarLogger.Infof("Reaction from %s in #%s: %s", r.UserID, channelName, r.Emoji.Name)

	_, err := service.CreateDiscordReaction(model.DiscordReaction{
		DiscordUserID: r.UserID,
		ChannelID:     r.ChannelID,
		ChannelName:   channelName,
		MessageID:     r.MessageID,
		Emoji:         r.Emoji.Name,
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to persist discord reaction: %v", err)
	}
}

func OnGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	logger.SugarLogger.Infof("GuildMemberAdd: user=%s roles=%v", m.User.ID, m.Roles)
}

func OnGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	if m.User == nil {
		return
	}
	if m.BeforeUpdate == nil {
		logger.SugarLogger.Infof("GuildMemberUpdate: user=%s roles=%v (no prior state)", m.User.ID, m.Roles)
	} else {
		added, removed := diffRoles(m.BeforeUpdate.Roles, m.Roles)
		if len(added) > 0 || len(removed) > 0 {
			logger.SugarLogger.Infof("GuildMemberUpdate: user=%s added=%v removed=%v", m.User.ID, added, removed)
		}
	}

	service.SyncDiscordUserAvatar(m.User.ID, m.AvatarURL("256"))
	if err := service.ReconcileGroupsForDiscordUser(m.User.ID, m.Roles); err != nil {
		logger.SugarLogger.Errorf("group sync: reconcile failed for %s: %v", m.User.ID, err)
	}
}

// OnReady fires when the Discord gateway is connected and the initial guild
// data is loaded. We kick a one-shot full reconcile to catch any drift that
// accumulated while the bot was offline (missed role changes, users who
// left the guild while we were down, etc). Reconnects also fire Ready; the
// sync.Once guard keeps this to a single sweep per process.
func OnReady(s *discordgo.Session, r *discordgo.Ready) {
	readyOnce.Do(func() {
		logger.SugarLogger.Infof("Discord gateway ready, kicking initial group sync")
		service.TriggerReconcileAll()
	})
}

// OnGuildMemberRemove fires when a user leaves (or is kicked/banned) the
// configured guild. Strip their DISCORD-sourced group memberships so
// access doesn't outlive guild membership. Reconcile with an empty role
// set computes a "desired = {}" and removes every DISCORD-sourced row in
// one diff — no separate code path required.
func OnGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.GuildID != config.DiscordGuild {
		return
	}
	logger.SugarLogger.Infof("GuildMemberRemove: user=%s, stripping DISCORD-sourced memberships", m.User.ID)
	if err := service.ReconcileGroupsForDiscordUser(m.User.ID, []string{}); err != nil {
		logger.SugarLogger.Errorf("group sync: leave-cleanup failed for %s: %v", m.User.ID, err)
	}
}

func OnUserUpdate(s *discordgo.Session, u *discordgo.UserUpdate) {
	if u == nil || u.User == nil {
		return
	}
	logger.SugarLogger.Infof("UserUpdate: user=%s username=%s", u.ID, u.Username)
	member, err := service.GetGuildMember(u.ID)
	if err != nil {
		logger.SugarLogger.Debugf("UserUpdate: user %s not in configured guild: %v", u.ID, err)
		return
	}
	service.SyncDiscordUserAvatar(u.ID, member.AvatarURL("256"))
}

// OnThreadUpdate keeps guild threads alive indefinitely. Discord doesn't
// allow disabling thread auto-archival (7-day window at most), so when a
// thread flips to archived we immediately flip it back. Unarchiving emits
// another ThreadUpdate with Archived=false, which falls through the guard
// below — no loop. Locked threads are left alone: locking is an explicit
// moderator "this thread is closed" signal, and force-unarchiving those
// would fight moderation.
func OnThreadUpdate(s *discordgo.Session, t *discordgo.ThreadUpdate) {
	if t.GuildID != config.DiscordGuild {
		return
	}
	if t.ThreadMetadata == nil || !t.ThreadMetadata.Archived || t.ThreadMetadata.Locked {
		return
	}
	logger.SugarLogger.Infof("ThreadUpdate: thread %s (%s) was archived, keeping alive", t.ID, t.Name)
	service.KeepThreadAlive(t.Channel)
}

func diffRoles(before, after []string) (added, removed []string) {
	beforeSet := make(map[string]struct{}, len(before))
	for _, r := range before {
		beforeSet[r] = struct{}{}
	}
	afterSet := make(map[string]struct{}, len(after))
	for _, r := range after {
		afterSet[r] = struct{}{}
	}
	for r := range afterSet {
		if _, ok := beforeSet[r]; !ok {
			added = append(added, r)
		}
	}
	for r := range beforeSet {
		if _, ok := afterSet[r]; !ok {
			removed = append(removed, r)
		}
	}
	return added, removed
}
