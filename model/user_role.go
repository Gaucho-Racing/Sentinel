package model

import "time"

type UserRole struct {
	UserID    string    `json:"user_id" gorm:"primaryKey"`
	Role      string    `json:"role" gorm:"primaryKey"`
	CreatedAt time.Time `json:"time" gorm:"autoCreateTime"`
}

func (UserRole) TableName() string {
	return "user_role"
}
