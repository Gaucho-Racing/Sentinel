package model

import "time"

type UserLogin struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	UserID      string    `json:"user_id"`
	Destination string    `json:"destination"`
	IPAddress   string    `json:"ip_address"`
	LoginType   string    `json:"login_type"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
