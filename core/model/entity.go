package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap is a free-form jsonb-backed map for storing arbitrary per-row
// metadata without growing a column every time a new provider hands us
// another field. Same Valuer/Scanner shape as StringSlice in group.go.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

type EntityType string

const (
	EntityTypeUser           EntityType = "USER"
	EntityTypeServiceAccount EntityType = "SERVICE_ACCOUNT"
)

type ExternalAuthProvider string

const (
	ExternalAuthProviderGoogle  ExternalAuthProvider = "GOOGLE"
	ExternalAuthProviderGitHub  ExternalAuthProvider = "GITHUB"
	ExternalAuthProviderDiscord ExternalAuthProvider = "DISCORD"
)

type Entity struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Type      EntityType `json:"type"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	EmailAuth      EntityEmail          `json:"email_auth" gorm:"-"`
	PhoneAuth      EntityPhone          `json:"phone_auth" gorm:"-"`
	ExternalAuths  []EntityExternalAuth `json:"external_auths" gorm:"-"`
	User           *User                `json:"user,omitempty" gorm:"-"`
	ServiceAccount *ServiceAccount      `json:"service_account,omitempty" gorm:"-"`
}

func (Entity) TableName() string {
	return "auth_entity"
}

type EntityEmail struct {
	EntityID  string    `json:"entity_id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"index"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (EntityEmail) TableName() string {
	return "auth_entity_email"
}

type EntityPhone struct {
	EntityID    string    `json:"entity_id" gorm:"primaryKey"`
	PhoneNumber string    `json:"phone_number" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (EntityPhone) TableName() string {
	return "auth_entity_phone"
}

type EntityExternalAuth struct {
	EntityID     string               `json:"entity_id" gorm:"primaryKey"`
	ExternalID   string               `json:"external_id"`
	Provider     ExternalAuthProvider `json:"provider" gorm:"primaryKey"`
	// Arbitrary per-provider data — email, username, avatar, etc. Provider
	// keys are whatever each provider hands back (Discord: email, username,
	// global_name, avatar, verified; Google would put its own shape here).
	// NOT the entity's primary login email (that's EntityEmail) — a user's
	// Discord / Google account email can intentionally differ. Refreshed on
	// every successful provider login so it stays current.
	Metadata     JSONMap              `json:"metadata" gorm:"type:jsonb"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresAt    time.Time            `json:"expires_at"`
	CreatedAt    time.Time            `json:"created_at" gorm:"autoCreateTime"`
}

func (EntityExternalAuth) TableName() string {
	return "auth_entity_external_auth"
}

