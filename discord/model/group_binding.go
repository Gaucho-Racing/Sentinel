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

// GroupDiscordRoleBinding ties a Sentinel group to a set of Discord role IDs.
// A user matches the binding only if they hold ALL DiscordRoleIDs on Discord
// (AND within a binding). Group membership is the OR across all bindings on
// the same group — so adding multiple bindings expresses "either-or" rules.
//
// Owned by the discord service: bindings are an integration-side concept and
// reference Sentinel group IDs from core but live in discord's domain.
type GroupDiscordRoleBinding struct {
	ID             string      `json:"id" gorm:"primaryKey"`
	GroupID        string      `json:"group_id" gorm:"index"`
	DiscordRoleIDs StringSlice `json:"discord_role_ids" gorm:"type:jsonb"`
	CreatedAt      time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

func (GroupDiscordRoleBinding) TableName() string {
	return "group_discord_role_binding"
}
