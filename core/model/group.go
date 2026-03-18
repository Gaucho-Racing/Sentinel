package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

type Group struct {
	ID             string      `json:"id" gorm:"primaryKey"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	AllowedSources StringSlice `json:"allowed_sources" gorm:"type:jsonb"`
	CreatedAt      time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

func (Group) TableName() string {
	return "group"
}

type GroupMember struct {
	GroupID       string    `json:"group_id" gorm:"primaryKey"`
	EntityID      string    `json:"entity_id" gorm:"primaryKey"`
	Source        string    `json:"source"`
	AddedBy       string    `json:"added_by"`
	HasExpiration bool      `json:"has_expiration"`
	ExpiresAt     time.Time `json:"expires_at"`
	JoinedAt      time.Time `json:"joined_at" gorm:"autoCreateTime"`
}

func (GroupMember) TableName() string {
	return "group_member"
}
