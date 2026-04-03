package model

import "time"

type DiscordMessage struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EntityID      string    `json:"entity_id"`
	DiscordUserID string    `json:"discord_user_id" gorm:"index"`
	ChannelID     string    `json:"channel_id"`
	ChannelName   string    `json:"channel_name"`
	MessageID     string    `json:"message_id"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DiscordMessage) TableName() string {
	return "discord_message"
}
