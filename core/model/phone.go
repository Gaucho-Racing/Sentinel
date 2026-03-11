package model

import "time"

type PhoneLoginCode struct {
	PhoneNumber string    `json:"phone_number" gorm:"primaryKey"`
	Code        int       `json:"code" gorm:"primaryKey"`
	ExpiresAt   time.Time `json:"expires_at"`
	Verified    bool      `json:"verified"`
	VerifiedAt  time.Time `json:"verified_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (PhoneLoginCode) TableName() string {
	return "auth_phone_login_code"
}
