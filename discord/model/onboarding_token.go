package model

import "time"

type OnboardingToken struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	DiscordID         string     `json:"discord_id" gorm:"index"`
	DiscordUsername   string     `json:"discord_username"`
	DiscordGlobalName string     `json:"discord_global_name"`
	DiscordAvatarURL  string     `json:"discord_avatar_url"`
	EntityID          string     `json:"entity_id" gorm:"index"`
	UsedAt            *time.Time `json:"used_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (OnboardingToken) TableName() string {
	return "onboarding_token"
}
