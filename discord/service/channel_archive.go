package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

func getChannel(channelID string) (*discordgo.Channel, error) {
	if ch, err := Discord.State.Channel(channelID); err == nil && ch != nil {
		return ch, nil
	}
	return Discord.Channel(channelID)
}

func IsArchiveCategory(channelID string) bool {
	ch, err := getChannel(channelID)
	if err != nil {
		logger.SugarLogger.Errorf("channel archive: failed to get channel %s: %v", channelID, err)
		return false
	}
	return ch.Type == discordgo.ChannelTypeGuildCategory && strings.EqualFold(ch.Name, config.DiscordArchiveCategoryName)
}

func GetArchivedChannel(channelID string) (model.ArchivedChannel, error) {
	var record model.ArchivedChannel
	if err := database.DB.Where("channel_id = ?", channelID).First(&record).Error; err != nil {
		return model.ArchivedChannel{}, err
	}
	return record, nil
}

// ArchiveChannel handles a channel that was just moved into the archive
// category: it snapshots the channel's own permission overwrites and previous
// parent, syncs the archive category's overwrites onto the channel, and posts
// a notice listing the roles that can still see it. If a snapshot already
// exists (channel was archived before and dragged back in), the original
// snapshot is kept so a later unarchive restores the true pre-archive state.
func ArchiveChannel(channel *discordgo.Channel, previousParentID string) {
	category, err := getChannel(channel.ParentID)
	if err != nil {
		logger.SugarLogger.Errorf("channel archive: failed to get archive category %s: %v", channel.ParentID, err)
		return
	}

	if _, err := GetArchivedChannel(channel.ID); err != nil {
		overwrites, err := json.Marshal(channel.PermissionOverwrites)
		if err != nil {
			logger.SugarLogger.Errorf("channel archive: failed to marshal overwrites for %s (%s): %v", channel.ID, channel.Name, err)
			return
		}
		record := model.ArchivedChannel{
			ChannelID:          channel.ID,
			ChannelName:        channel.Name,
			PreviousParentID:   previousParentID,
			PreviousOverwrites: string(overwrites),
		}
		if err := database.DB.Create(&record).Error; err != nil {
			logger.SugarLogger.Errorf("channel archive: failed to persist snapshot for %s (%s): %v", channel.ID, channel.Name, err)
			return
		}
	} else {
		logger.SugarLogger.Infof("channel archive: snapshot already exists for %s (%s), keeping original", channel.ID, channel.Name)
	}

	if len(category.PermissionOverwrites) > 0 {
		_, err = Discord.ChannelEdit(channel.ID, &discordgo.ChannelEdit{
			PermissionOverwrites: category.PermissionOverwrites,
		})
		if err != nil {
			logger.SugarLogger.Errorf("channel archive: failed to sync category permissions onto %s (%s): %v", channel.ID, channel.Name, err)
			return
		}
	}
	logger.SugarLogger.Infof("channel archive: archived channel %s (%s)", channel.ID, channel.Name)

	content := fmt.Sprintf("This channel has been archived and is now only visible to %s. Run `%sunarchive` to restore it.",
		strings.Join(viewerMentions(category.PermissionOverwrites), " "), config.DiscordPrefix)
	sendMessageWithoutPings(channel.ID, content)
}

// UnarchiveChannel moves an archived channel back to its previous category
// and restores its own permission overwrites from the snapshot. Returns the
// consumed snapshot so callers can report what was restored.
func UnarchiveChannel(channelID string) (model.ArchivedChannel, error) {
	record, err := GetArchivedChannel(channelID)
	if err != nil {
		return model.ArchivedChannel{}, fmt.Errorf("channel %s is not archived", channelID)
	}

	var overwrites []*discordgo.PermissionOverwrite
	if record.PreviousOverwrites != "" {
		if err := json.Unmarshal([]byte(record.PreviousOverwrites), &overwrites); err != nil {
			return record, fmt.Errorf("failed to unmarshal overwrite snapshot: %w", err)
		}
	}

	// ChannelEdit's ParentID and PermissionOverwrites are omitempty, so a
	// snapshot with no parent or no overwrites can't be expressed in a single
	// edit: the parent is left as-is (caller surfaces this), and an empty
	// overwrite set is restored by deleting the category-synced overwrites
	// individually below.
	edit := &discordgo.ChannelEdit{ParentID: record.PreviousParentID}
	if len(overwrites) > 0 {
		edit.PermissionOverwrites = overwrites
	}
	if _, err := Discord.ChannelEdit(channelID, edit); err != nil {
		return record, fmt.Errorf("failed to restore channel: %w", err)
	}
	if len(overwrites) == 0 {
		channel, err := Discord.Channel(channelID)
		if err != nil {
			return record, fmt.Errorf("failed to get channel for overwrite cleanup: %w", err)
		}
		for _, overwrite := range channel.PermissionOverwrites {
			if err := Discord.ChannelPermissionDelete(channelID, overwrite.ID); err != nil {
				logger.SugarLogger.Errorf("channel archive: failed to delete overwrite %s on %s: %v", overwrite.ID, channelID, err)
			}
		}
	}

	if err := database.DB.Delete(&record).Error; err != nil {
		logger.SugarLogger.Errorf("channel archive: failed to delete snapshot for %s: %v", channelID, err)
	}
	logger.SugarLogger.Infof("channel archive: unarchived channel %s (%s)", channelID, record.ChannelName)
	return record, nil
}

// viewerMentions returns mention strings for the roles and members granted
// VIEW_CHANNEL by the given overwrites.
func viewerMentions(overwrites []*discordgo.PermissionOverwrite) []string {
	var mentions []string
	for _, overwrite := range overwrites {
		if overwrite.Allow&discordgo.PermissionViewChannel == 0 {
			continue
		}
		switch overwrite.Type {
		case discordgo.PermissionOverwriteTypeRole:
			mentions = append(mentions, "<@&"+overwrite.ID+">")
		case discordgo.PermissionOverwriteTypeMember:
			mentions = append(mentions, "<@"+overwrite.ID+">")
		}
	}
	return mentions
}

// sendMessageWithoutPings posts a message whose role/user mentions render but
// don't notify anyone (zero-value AllowedMentions suppresses all pings).
func sendMessageWithoutPings(channelID, content string) {
	_, err := Discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:         content,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
	if err != nil {
		logger.SugarLogger.Errorf("Failed to send message in %s: %v", channelID, err)
	}
}
