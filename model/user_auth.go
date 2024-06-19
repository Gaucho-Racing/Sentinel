package model

import "time"

type UserAuth struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email" gorm:"index"`
	Password  string    `json:"password"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserAuth) TableName() string {
	return "user_auth"
}
