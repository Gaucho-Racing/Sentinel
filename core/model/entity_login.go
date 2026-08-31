package model

import "time"

type EntityLogin struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	EntityID       string    `json:"entity_id" gorm:"index;index:idx_entity_login_entity_created,priority:1"`
	ClientID       string    `json:"client_id" gorm:"index;index:idx_entity_login_client_created,priority:1"`
	Scope          string    `json:"scope"`
	AccessTokenID  string    `json:"access_token_id"`
	RefreshTokenID string    `json:"refresh_token_id"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_entity_login_created_at;index:idx_entity_login_entity_created,priority:2;index:idx_entity_login_client_created,priority:2"`
}

func (EntityLogin) TableName() string {
	return "entity_login"
}
