package model

import (
	"time"
)

var ValidOauthScopes = map[string]string{
	"user:read":         "Read user account information",
	"user:write":        "Edit user account information",
	"drive:read":        "Read user's team drive access information",
	"drive:write":       "Add/remove user from the team drive",
	"github:read":       "Read user's github access information",
	"github:write":      "Add/remove user from the github org",
	"applications:read": "Read user's applications (this includes the client id and secret)",
	"logins:read":       "Read users's login history",
	"sentinel:all":      "Internal scope for Sentinel, client applications should not request this scope.",
}

type ClientApplication struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	UserID       string    `json:"user_id"`
	Secret       string    `json:"secret"`
	Name         string    `json:"name"`
	RedirectURIs []string  `json:"redirect_uris" gorm:"-"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ClientApplication) TableName() string {
	return "client_application"
}

type ClientApplicationRedirectURI struct {
	ClientApplicationID string `gorm:"primaryKey" json:"client_application_id"`
	RedirectURI         string `gorm:"primaryKey" json:"redirect_uri"`
}

func (ClientApplicationRedirectURI) TableName() string {
	return "client_application_redirect_uri"
}

type AuthorizationCode struct {
	Code      string    `gorm:"primaryKey" json:"code"`
	ClientID  string    `json:"client_id"`
	UserID    string    `json:"user_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AuthorizationCode) TableName() string {
	return "authorization_code"
}

type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
