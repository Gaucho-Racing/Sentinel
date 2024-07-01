package model

import "time"

type ClientApplication struct {
	ID           string    `json:"id"`
	Secret       string    `json:"secret"`
	Name         string    `json:"name"`
	RedirectURIs []string  `json:"redirect_uris"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ClientApplication) TableName() string {
	return "client_application"
}

type ClientApplicationRedirectURI struct {
	ClientApplicationID string `json:"client_application_id"`
	RedirectURI         string `json:"redirect_uri"`
}

func (ClientApplicationRedirectURI) TableName() string {
	return "client_application_redirect_uri"
}
