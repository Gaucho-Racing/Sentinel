package model

import "time"

type ServiceAccount struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EntityID      string    `json:"entity_id" gorm:"index"`
	ApplicationID string    `json:"application_id" gorm:"index"`
	Name          string    `json:"name"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ServiceAccount) TableName() string {
	return "service_account"
}
