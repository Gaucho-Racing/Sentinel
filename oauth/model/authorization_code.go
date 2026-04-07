package model

import "time"

type AuthorizationCode struct {
	Code        string    `json:"code" gorm:"primaryKey"`
	EntityID    string    `json:"entity_id"`
	ClientID    string    `json:"client_id"`
	Scope       string    `json:"scope"`
	RedirectURI string    `json:"redirect_uri"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (AuthorizationCode) TableName() string {
	return "authorization_code"
}
