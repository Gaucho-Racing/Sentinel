package model

import "time"

type Application struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	OwnerID      string    `json:"owner_id" gorm:"index"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ClientID     string    `json:"client_id" gorm:"uniqueIndex"`
	ClientSecret string    `json:"-"`
	IconURL      string    `json:"icon_url"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Application) TableName() string {
	return "application"
}
