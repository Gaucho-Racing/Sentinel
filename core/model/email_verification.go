package model

import "time"

type EmailVerification struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	EntityID    string     `json:"entity_id" gorm:"index"`
	Purpose     string     `json:"purpose" gorm:"index;default:verify_email"`
	Destination string     `json:"destination" gorm:"column:email"`
	CodeHash    string     `json:"-"`
	Attempts    int        `json:"attempts"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ConsumedAt  *time.Time `json:"consumed_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
}

func (EmailVerification) TableName() string {
	return "auth_email_verification"
}
