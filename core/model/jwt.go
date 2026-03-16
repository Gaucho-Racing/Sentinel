package model

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	Scope        string                 `json:"scope"`
	CustomClaims map[string]interface{} `json:"-"`
	jwt.RegisteredClaims
}

var registeredKeys = map[string]bool{
	"jti": true, "sub": true, "iss": true,
	"aud": true, "exp": true, "iat": true, "nbf": true,
	"scope": true,
}

func (tc TokenClaims) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for k, v := range tc.CustomClaims {
		m[k] = v
	}
	m["scope"] = tc.Scope
	if tc.ID != "" {
		m["jti"] = tc.ID
	}
	if tc.Subject != "" {
		m["sub"] = tc.Subject
	}
	if tc.Issuer != "" {
		m["iss"] = tc.Issuer
	}
	if len(tc.Audience) > 0 {
		m["aud"] = tc.Audience
	}
	if tc.ExpiresAt != nil {
		m["exp"] = tc.ExpiresAt
	}
	if tc.IssuedAt != nil {
		m["iat"] = tc.IssuedAt
	}
	if tc.NotBefore != nil {
		m["nbf"] = tc.NotBefore
	}
	return json.Marshal(m)
}

func (tc *TokenClaims) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &tc.RegisteredClaims); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if scopeRaw, ok := raw["scope"]; ok {
		json.Unmarshal(scopeRaw, &tc.Scope)
	}
	tc.CustomClaims = make(map[string]interface{})
	for k, v := range raw {
		if !registeredKeys[k] {
			var val interface{}
			json.Unmarshal(v, &val)
			tc.CustomClaims[k] = val
		}
	}
	return nil
}

type Token struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	EntityID  string    `json:"entity_id"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Token) TableName() string {
	return "auth_token"
}
