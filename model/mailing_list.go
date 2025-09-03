package model

import "time"

type MailingList struct {
	Email        string    `gorm:"primaryKey" json:"email" binding:"required,email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         string    `json:"role"`
	Organization string    `json:"organization"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MailingList) TableName() string {
	return "mailing_list"
}
