package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SCIMSyncStatus string

const (
	SCIMSyncStatusNever     SCIMSyncStatus = "NEVER"
	SCIMSyncStatusQueued    SCIMSyncStatus = "QUEUED"
	SCIMSyncStatusRunning   SCIMSyncStatus = "RUNNING"
	SCIMSyncStatusSucceeded SCIMSyncStatus = "SUCCEEDED"
	SCIMSyncStatusFailed    SCIMSyncStatus = "FAILED"
)

type SCIMSyncInterval string

const (
	SCIMSyncInterval5Minutes  SCIMSyncInterval = "5m"
	SCIMSyncInterval15Minutes SCIMSyncInterval = "15m"
	SCIMSyncInterval30Minutes SCIMSyncInterval = "30m"
	SCIMSyncInterval1Hour     SCIMSyncInterval = "1h"
	SCIMSyncInterval6Hours    SCIMSyncInterval = "6h"
	SCIMSyncIntervalDaily     SCIMSyncInterval = "24h"
)

type SCIMConfiguration struct {
	ApplicationID  string           `json:"application_id" gorm:"primaryKey"`
	Endpoint       string           `json:"endpoint"`
	AccessToken    string           `json:"-"`
	TokenExpiresAt *time.Time       `json:"token_expires_at"`
	Enabled        bool             `json:"enabled"`
	SyncInterval   SCIMSyncInterval `json:"sync_interval"`
	NextSyncAt     *time.Time       `json:"next_sync_at" gorm:"index"`
	LastSyncAt     *time.Time       `json:"last_sync_at"`
	LastSyncStatus SCIMSyncStatus   `json:"last_sync_status"`
	LastSyncError  string           `json:"last_sync_error"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
}

type SCIMSyncRunStatus string

const (
	SCIMSyncRunStatusQueued    SCIMSyncRunStatus = "QUEUED"
	SCIMSyncRunStatusRunning   SCIMSyncRunStatus = "RUNNING"
	SCIMSyncRunStatusSucceeded SCIMSyncRunStatus = "SUCCEEDED"
	SCIMSyncRunStatusFailed    SCIMSyncRunStatus = "FAILED"
)

type SCIMSyncTrigger string

const (
	SCIMSyncTriggerManual    SCIMSyncTrigger = "MANUAL"
	SCIMSyncTriggerScheduled SCIMSyncTrigger = "SCHEDULED"
)

type SCIMSkippedUser struct {
	EntityID string   `json:"entity_id"`
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
	Reason   string   `json:"reason"`
}

type SCIMSkippedUsers []SCIMSkippedUser

func (s SCIMSkippedUsers) Value() (driver.Value, error) {
	value, err := json.Marshal(s)
	return string(value), err
}

func (s *SCIMSkippedUsers) Scan(value any) error {
	return scanSCIMJSON(value, s)
}

type SCIMValidationErrors []string

func (e SCIMValidationErrors) Value() (driver.Value, error) {
	value, err := json.Marshal(e)
	return string(value), err
}

func (e *SCIMValidationErrors) Scan(value any) error {
	return scanSCIMJSON(value, e)
}

func scanSCIMJSON(value, destination any) error {
	switch typed := value.(type) {
	case string:
		return json.Unmarshal([]byte(typed), destination)
	case []byte:
		return json.Unmarshal(typed, destination)
	case nil:
		return nil
	default:
		return fmt.Errorf("unsupported SCIM JSON type: %T", value)
	}
}

type SCIMSyncRun struct {
	ID                 string               `json:"id" gorm:"primaryKey"`
	ApplicationID      string               `json:"application_id" gorm:"index:idx_scim_run_application_requested,priority:1"`
	Trigger            SCIMSyncTrigger      `json:"trigger"`
	Status             SCIMSyncRunStatus    `json:"status" gorm:"index:idx_scim_run_status_requested,priority:1"`
	DesiredUsers       int                  `json:"desired_users"`
	DesiredGroups      int                  `json:"desired_groups"`
	UsersCreated       int                  `json:"users_created"`
	UsersUpdated       int                  `json:"users_updated"`
	UsersDeactivated   int                  `json:"users_deactivated"`
	GroupsCreated      int                  `json:"groups_created"`
	GroupsUpdated      int                  `json:"groups_updated"`
	MembershipsAdded   int                  `json:"memberships_added"`
	MembershipsRemoved int                  `json:"memberships_removed"`
	SkippedUsers       SCIMSkippedUsers     `json:"skipped_users" gorm:"type:jsonb"`
	ValidationErrors   SCIMValidationErrors `json:"validation_errors" gorm:"type:jsonb"`
	Error              string               `json:"error"`
	RequestedAt        time.Time            `json:"requested_at" gorm:"autoCreateTime;index:idx_scim_run_application_requested,priority:2,sort:desc;index:idx_scim_run_status_requested,priority:2"`
	StartedAt          *time.Time           `json:"started_at"`
	CompletedAt        *time.Time           `json:"completed_at"`
	LeaseExpiresAt     *time.Time           `json:"-" gorm:"index"`
}

func (SCIMSyncRun) TableName() string {
	return "saml_scim_sync_run"
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
