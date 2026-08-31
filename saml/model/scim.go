package model

import "time"

type SCIMSyncStatus string

const (
	SCIMSyncStatusNever     SCIMSyncStatus = "NEVER"
	SCIMSyncStatusRunning   SCIMSyncStatus = "RUNNING"
	SCIMSyncStatusSucceeded SCIMSyncStatus = "SUCCEEDED"
	SCIMSyncStatusFailed    SCIMSyncStatus = "FAILED"
)

type SCIMConfiguration struct {
	ApplicationID  string         `json:"application_id" gorm:"primaryKey"`
	Endpoint       string         `json:"endpoint"`
	AccessToken    string         `json:"-"`
	TokenExpiresAt *time.Time     `json:"token_expires_at"`
	Enabled        bool           `json:"enabled"`
	LastSyncAt     *time.Time     `json:"last_sync_at"`
	LastSyncStatus SCIMSyncStatus `json:"last_sync_status"`
	LastSyncError  string         `json:"last_sync_error"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (SCIMConfiguration) TableName() string {
	return "saml_scim_configuration"
}

type SCIMResourceType string

const (
	SCIMResourceTypeUser  SCIMResourceType = "USER"
	SCIMResourceTypeGroup SCIMResourceType = "GROUP"
)

type SCIMResource struct {
	ApplicationID string           `gorm:"primaryKey"`
	ResourceType  SCIMResourceType `gorm:"primaryKey"`
	SentinelID    string           `gorm:"primaryKey"`
	ProviderID    string
	LastSyncedAt  time.Time
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (SCIMResource) TableName() string {
	return "saml_scim_resource"
}
