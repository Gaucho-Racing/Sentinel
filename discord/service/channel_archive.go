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

// archiveWriteMask covers the permissions stripped when a channel is
// archived: posting, threads, reactions, and voice connect. View and
// read-history bits are never touched, so an archived channel stays visible
// to exactly the audience that could see it before — just read-only.
const archiveWriteMask = discordgo.PermissionSendMessages |
	discordgo.PermissionSendMessagesInThreads |
	discordgo.PermissionCreatePublicThreads |
	discordgo.PermissionCreatePrivateThreads |
	discordgo.PermissionAddReactions |
	discordgo.PermissionVoiceConnect

// archiveExemptRoleIDs keep full access on archived channels. Roles with
// Administrator (Admin, Officer, etc.) bypass channel overwrites entirely
// and need no exemption here.
var archiveExemptRoleIDs = []string{
	config.RobotDiscordRoleID,
	config.DevOpsDiscordRoleID,
}

func getChannel(channelID string) (*discordgo.Channel, error) {
	if ch, err := Discord.State.Channel(channelID); err == nil && ch != nil {
		return ch, nil
	}
	return Discord.Channel(channelID)
}

// discordCategoryChannelCap is Discord's hard limit on channels per category.
const discordCategoryChannelCap = 50

// findOrCreateArchiveCategory returns an archive category with room for one
// more channel. Discord allows duplicate category names, so when every
// existing ARCHIVE category is at the cap (or none exists) a new one is
// provisioned at the bottom of the channel list, cloning permission
// overwrites from the last existing archive category when there is one.
func findOrCreateArchiveCategory() (*discordgo.Channel, error) {
	channels, err := GetGuildChannels()
	if err != nil {
		return nil, err
	}
	var archiveCategories []*discordgo.Channel
	childCounts := make(map[string]int)
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildCategory {
			if strings.EqualFold(ch.Name, config.DiscordArchiveCategoryName) {
				archiveCategories = append(archiveCategories, ch)
			}
		} else if ch.ParentID != "" {
			childCounts[ch.ParentID]++
		}
	}
	for _, category := range archiveCategories {
		if childCounts[category.ID] < discordCategoryChannelCap {
			return category, nil
		}
	}

	data := discordgo.GuildChannelCreateData{
		Name: config.DiscordArchiveCategoryName,
		Type: discordgo.ChannelTypeGuildCategory,
	}
	if len(archiveCategories) > 0 {
		data.PermissionOverwrites = archiveCategories[len(archiveCategories)-1].PermissionOverwrites
	} else {
		data.PermissionOverwrites = defaultArchiveCategoryOverwrites()
	}
	category, err := Discord.GuildChannelCreateComplex(config.DiscordGuild, data)
	if err != nil {
		return nil, fmt.Errorf("failed to provision new archive category: %w", err)
	}
	logger.SugarLogger.Infof("channel archive: provisioned new archive category %s (existing ones full: %d)", category.ID, len(archiveCategories))
	return category, nil
}

// defaultArchiveCategoryOverwrites is only used when provisioning the very
// first archive category: hidden from @everyone, visible to the exempt roles.
func defaultArchiveCategoryOverwrites() []*discordgo.PermissionOverwrite {
	overwrites := []*discordgo.PermissionOverwrite{{
		ID:   config.DiscordGuild,
		Type: discordgo.PermissionOverwriteTypeRole,
		Deny: discordgo.PermissionViewChannel,
	}}
	for _, roleID := range archiveExemptRoleIDs {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID:    roleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel,
		})
	}
	return overwrites
}

func GetArchivedChannel(channelID string) (model.ArchivedChannel, error) {
	var record model.ArchivedChannel
	if err := database.DB.Where("channel_id = ?", channelID).First(&record).Error; err != nil {
		return model.ArchivedChannel{}, err
	}
	return record, nil
}

func GetAllArchivedChannels() ([]model.ArchivedChannel, error) {
	var records []model.ArchivedChannel
	if err := database.DB.Order("archived_at desc").Find(&records).Error; err != nil {
		return []model.ArchivedChannel{}, err
	}
	return records, nil
}

// ArchiveChannel snapshots the channel's permission overwrites and parent
// category, moves it into the archive category, and rewrites its permissions
// to the standardized archived form (read-only for its existing audience,
// full access for the exempt roles). Posts a notice in the channel on
// success.
func ArchiveChannel(channelID, archivedByDiscordID string) error {
	if _, err := GetArchivedChannel(channelID); err == nil {
		return fmt.Errorf("channel %s is already archived", channelID)
	}
	channel, err := getChannel(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	category, err := findOrCreateArchiveCategory()
	if err != nil {
		return err
	}

	snapshot, err := json.Marshal(channel.PermissionOverwrites)
	if err != nil {
		return fmt.Errorf("failed to marshal overwrite snapshot: %w", err)
	}
	record := model.ArchivedChannel{
		ChannelID:           channel.ID,
		ChannelName:         channel.Name,
		PreviousParentID:    channel.ParentID,
		PreviousOverwrites:  string(snapshot),
		ArchivedByEntityID:  GetEntityIDForDiscordUser(archivedByDiscordID),
		ArchivedByDiscordID: archivedByDiscordID,
	}
	if err := database.DB.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to persist snapshot: %w", err)
	}

	_, err = Discord.ChannelEdit(channel.ID, &discordgo.ChannelEdit{
		ParentID:             category.ID,
		PermissionOverwrites: archivedOverwrites(channel.PermissionOverwrites),
	})
	if err != nil {
		// Roll back the snapshot so a retry doesn't hit "already archived".
		database.DB.Delete(&record)
		return fmt.Errorf("failed to move and lock channel: %w", err)
	}

	logger.SugarLogger.Infof("channel archive: archived channel %s (%s) by %s", channel.ID, channel.Name, archivedByDiscordID)
	sendMessageWithoutPings(channel.ID, fmt.Sprintf("This channel has been archived by <@%s> and is now read-only. Run `%sunarchive` to restore it.", archivedByDiscordID, config.DiscordPrefix))
	return nil
}

// UnarchiveChannel moves an archived channel back to its previous category
// and restores its snapshotted permission overwrites. Returns the consumed
// snapshot so callers can report what was restored.
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
	// overwrite set is restored by deleting the archive overwrites
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
		// Keep the snapshot if any deletion fails so the command can be
		// retried — the restore is idempotent, a retry just re-applies the
		// parent edit and deletes the remaining archive overwrites.
		for _, overwrite := range channel.PermissionOverwrites {
			if err := Discord.ChannelPermissionDelete(channelID, overwrite.ID); err != nil {
				return record, fmt.Errorf("failed to delete overwrite %s: %w", overwrite.ID, err)
			}
		}
	}

	if err := database.DB.Delete(&record).Error; err != nil {
		logger.SugarLogger.Errorf("channel archive: failed to delete snapshot for %s: %v", channelID, err)
	}
	logger.SugarLogger.Infof("channel archive: unarchived channel %s (%s)", channelID, record.ChannelName)
	return record, nil
}

// archivedOverwrites transforms a channel's overwrites into their archived
// form: every existing overwrite loses its write-bit allows, @everyone gets
// an explicit write deny (covering members whose write access comes from
// base permissions rather than an overwrite), and the exempt roles get view
// plus full write access back.
func archivedOverwrites(existing []*discordgo.PermissionOverwrite) []*discordgo.PermissionOverwrite {
	overwrites := make([]*discordgo.PermissionOverwrite, 0, len(existing)+len(archiveExemptRoleIDs)+1)
	index := make(map[string]*discordgo.PermissionOverwrite, len(existing))
	for _, overwrite := range existing {
		copied := *overwrite
		copied.Allow &^= archiveWriteMask
		overwrites = append(overwrites, &copied)
		index[copied.ID] = &copied
	}

	if everyone, ok := index[config.DiscordGuild]; ok {
		everyone.Deny |= archiveWriteMask
	} else {
		everyone = &discordgo.PermissionOverwrite{
			ID:   config.DiscordGuild,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: archiveWriteMask,
		}
		overwrites = append(overwrites, everyone)
		index[everyone.ID] = everyone
	}

	exemptMask := int64(archiveWriteMask | discordgo.PermissionViewChannel)
	for _, roleID := range archiveExemptRoleIDs {
		if exempt, ok := index[roleID]; ok {
			exempt.Allow |= exemptMask
			exempt.Deny &^= exemptMask
		} else {
			exempt = &discordgo.PermissionOverwrite{
				ID:    roleID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: exemptMask,
			}
			overwrites = append(overwrites, exempt)
			index[exempt.ID] = exempt
		}
	}
	return overwrites
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
