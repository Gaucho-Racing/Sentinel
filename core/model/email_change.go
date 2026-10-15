package model

import "time"

type EmailChange struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	EntityID  string    `json:"entity_id" gorm:"index"`
	OldEmail  string    `json:"old_email" gorm:"index"`
	NewEmail  string    `json:"new_email"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (EmailChange) TableName() string {
	return "auth_email_change"
}
