package model

import "time"

type MailingList struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email" binding:"required,email"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (MailingList) TableName() string {
	return "mailing_list"
}
