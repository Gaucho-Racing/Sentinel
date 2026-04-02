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

type GroupMemberSource string

const (
	GroupMemberSourceDirect      GroupMemberSource = "DIRECT"
	GroupMemberSourceConditional GroupMemberSource = "CONDITIONAL"
	GroupMemberSourceDiscord     GroupMemberSource = "DISCORD"
)

type GroupJoinRequestStatus string

const (
	GroupJoinRequestStatusPending  GroupJoinRequestStatus = "PENDING"
	GroupJoinRequestStatusApproved GroupJoinRequestStatus = "APPROVED"
	GroupJoinRequestStatusRejected GroupJoinRequestStatus = "REJECTED"
)

type Group struct {
	ID             string      `json:"id" gorm:"primaryKey"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	AllowedSources StringSlice `json:"allowed_sources" gorm:"type:jsonb"`
	UpdatedAt      time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
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

type GroupOwner struct {
	GroupID   string    `json:"group_id" gorm:"primaryKey"`
	EntityID  string    `json:"entity_id" gorm:"primaryKey"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (GroupOwner) TableName() string {
	return "group_owner"
}

type GroupJoinRequest struct {
	ID            string                    `json:"id" gorm:"primaryKey"`
	GroupID       string                    `json:"group_id"`
	EntityID      string                    `json:"entity_id"`
	Status        string                    `json:"status"`
	ReviewedBy    string                    `json:"reviewed_by"`
	ReviewedAt    time.Time                 `json:"reviewed_at"`
	HasExpiration bool                      `json:"has_expiration"`
	ExpiresAt     time.Time                 `json:"expires_at"`
	CreatedAt     time.Time                 `json:"created_at" gorm:"autoCreateTime"`
	Comments      []GroupJoinRequestComment `json:"comments" gorm:"-"`
}

func (GroupJoinRequest) TableName() string {
	return "group_join_request"
}

type GroupJoinRequestComment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	RequestID string    `json:"request_id"`
	EntityID  string    `json:"entity_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (GroupJoinRequestComment) TableName() string {
	return "group_join_request_comment"
}
