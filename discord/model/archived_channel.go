package model

import "time"

// ArchivedChannel snapshots a channel's pre-archive state so it can be
// restored by the unarchive command. PreviousOverwrites holds the channel's
// own permission overwrites as JSON ([]*discordgo.PermissionOverwrite).
type ArchivedChannel struct {
	ChannelID          string    `json:"channel_id" gorm:"primaryKey"`
	ChannelName        string    `json:"channel_name"`
	PreviousParentID   string    `json:"previous_parent_id"`
	PreviousOverwrites string    `json:"previous_overwrites"`
	ArchivedBy         string    `json:"archived_by"`
	ArchivedAt         time.Time `json:"archived_at" gorm:"autoCreateTime"`
}

func (ArchivedChannel) TableName() string {
	return "archived_channel"
}
