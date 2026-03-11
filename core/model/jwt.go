package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	Entity Entity `json:"entity,omitempty"`
	Scope  string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	Scope string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

type RefreshToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	EntityID  string    `json:"entity_id"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RefreshToken) TableName() string {
	return "auth_refresh_token"
}
