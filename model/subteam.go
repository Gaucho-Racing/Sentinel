package model

import "time"

type Subteam struct {
	RoleID    string    `json:"role_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Subteam) TableName() string {
	return "subteam"
}
