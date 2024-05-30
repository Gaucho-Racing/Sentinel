package model

import "time"

type UserActivity struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserActivity) TableName() string {
	return "user_activity"
}
