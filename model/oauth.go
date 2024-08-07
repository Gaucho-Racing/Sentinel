package model

import (
	"time"
)

var ValidOauthScopes = []string{
	"read:user",
	"write:user",
	"read:drive",
	"write:drive",
	"read:github",
	"write:github",
	"read:applications",
	"read:logins",
	"sentinel:all",
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
