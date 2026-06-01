package model

import "time"

type Application struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	OwnerID      string    `json:"owner_id" gorm:"index"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ClientID     string    `json:"client_id" gorm:"uniqueIndex"`
	ClientSecret string    `json:"-"`
	IconURL      string    `json:"icon_url"`
	LaunchURL    string    `json:"launch_url"`
	RedirectURIs []string  `json:"redirect_uris" gorm:"-"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Application) TableName() string {
	return "application"
}

type ApplicationRedirectURI struct {
	ApplicationID string `json:"application_id" gorm:"primaryKey"`
	RedirectURI   string `json:"redirect_uri" gorm:"primaryKey"`
}

func (ApplicationRedirectURI) TableName() string {
	return "application_redirect_uri"
}

// ApplicationGroup links an application to a group. The Required flag, when
// true, gates access to the application: a user must hold membership in at
// least one Required-flagged group on the app to obtain a token for it.
// Non-required links still flow into the token's groups claim — they just
// don't block access.
type ApplicationGroup struct {
	ApplicationID string    `json:"application_id" gorm:"primaryKey"`
	GroupID       string    `json:"group_id" gorm:"primaryKey"`
	Required      bool      `json:"required"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ApplicationGroup) TableName() string {
	return "application_group"
}
