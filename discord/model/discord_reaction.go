package model

import "time"

type DiscordReaction struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EntityID      string    `json:"entity_id"`
	DiscordUserID string    `json:"discord_user_id" gorm:"index"`
	ChannelID     string    `json:"channel_id"`
	ChannelName   string    `json:"channel_name"`
	MessageID     string    `json:"message_id"`
	Emoji         string    `json:"emoji"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DiscordReaction) TableName() string {
	return "discord_reaction"
}
