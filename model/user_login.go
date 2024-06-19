package model

import "time"

type UserLogin struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email" gorm:"index"`
	Password  string    `json:"password"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserLogin) TableName() string {
	return "user_login"
}
