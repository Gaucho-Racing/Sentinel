package model

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type TokenResponse struct {
	IDToken      string `json:"id_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type AuthClaims struct {
	Name           string   `json:"name,omitempty"`
	GivenName      string   `json:"given_name,omitempty"`
	FamilyName     string   `json:"family_name,omitempty"`
	Profile        string   `json:"profile,omitempty"`
	Picture        string   `json:"picture,omitempty"`
	Email          string   `json:"email,omitempty"`
	EmailVerified  bool     `json:"email_verified,omitempty"`
	BookstackRoles []string `json:"bookstack_roles,omitempty"`
	Scope          string   `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

func (c AuthClaims) Valid() error {
	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc()

	if !c.VerifyExpiresAt(now, true) {
		delta := now.Sub(c.ExpiresAt.Time)
		vErr.Inner = fmt.Errorf("%s by %s", jwt.ErrTokenExpired, delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	if !c.VerifyIssuedAt(now, true) {
		vErr.Inner = jwt.ErrTokenUsedBeforeIssued
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}

	if !c.VerifyIssuer("https://sso.gauchoracing.com", true) {
		vErr.Inner = jwt.ErrTokenInvalidIssuer
		vErr.Errors |= jwt.ValidationErrorIssuer
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}
