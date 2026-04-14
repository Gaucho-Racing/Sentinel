package model

import "time"

type EntityLogin struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	EntityID       string    `json:"entity_id" gorm:"index"`
	ClientID       string    `json:"client_id" gorm:"index"`
	Scope          string    `json:"scope"`
	AccessTokenID  string    `json:"access_token_id"`
	RefreshTokenID string    `json:"refresh_token_id"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (EntityLogin) TableName() string {
	return "entity_login"
}
