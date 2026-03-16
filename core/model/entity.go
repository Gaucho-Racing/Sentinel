package model

import "time"

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
	Type      string    `json:"type"`
	OwnedBy   string    `json:"owned_by"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	EmailAuth     EntityEmail          `json:"email_auth" gorm:"-"`
	PhoneAuth     EntityPhone          `json:"phone_auth" gorm:"-"`
	ExternalAuths []EntityExternalAuth `json:"external_auths" gorm:"-"`
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
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresAt    time.Time            `json:"expires_at"`
	CreatedAt    time.Time            `json:"created_at" gorm:"autoCreateTime"`
}

func (EntityExternalAuth) TableName() string {
	return "auth_entity_external_auth"
}

type User struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	Username              string    `json:"username"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email"`
	PhoneNumber           string    `json:"phone_number"`
	Gender                string    `json:"gender"`
	Birthday              string    `json:"birthday"`
	GraduateLevel         string    `json:"graduate_level"`
	GraduationYear        int       `json:"graduation_year"`
	Major                 string    `json:"major"`
	ShirtSize             string    `json:"shirt_size"`
	JacketSize            string    `json:"jacket_size"`
	SAERegistrationNumber string    `json:"sae_registration_number"`
	AvatarURL             string    `json:"avatar_url"`
	UpdatedAt             time.Time `json:"updated_at"`
	CreatedAt             time.Time `json:"created_at"`
}
