package model

import "time"

type EmailLoginCode struct {
	Email      string    `json:"email" gorm:"primaryKey"`
	Code       int       `json:"code" gorm:"primaryKey"`
	ExpiresAt  time.Time `json:"expires_at"`
	Verified   bool      `json:"verified"`
	VerifiedAt time.Time `json:"verified_at"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (EmailLoginCode) TableName() string {
	return "auth_email_login_code"
}
