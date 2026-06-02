package service

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
)

type oidcEntity struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	EmailAuth struct {
		Email string `json:"email"`
	} `json:"email_auth"`
	User *struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

func fetchOIDCEntity(entityID string) (oidcEntity, error) {
	var e oidcEntity
	if err := sentinel.Get("/core/entity/"+entityID, &e); err != nil {
		logger.SugarLogger.Errorf("Failed to load entity %s for OIDC claims: %v", entityID, err)
		return oidcEntity{}, err
	}
	return e, nil
}

// identityClaims maps the entity's identity onto the standard OIDC claims the
// granted scopes allow. `openid` on its own yields no profile/email claims —
// those require the `profile` and `email` scopes respectively.
func identityClaims(e oidcEntity, scope string) map[string]interface{} {
	claims := map[string]interface{}{}

	if ScopesContain(scope, "profile") && e.User != nil {
		if name := strings.TrimSpace(e.User.FirstName + " " + e.User.LastName); name != "" {
			claims["name"] = name
		}
		if e.User.FirstName != "" {
			claims["given_name"] = e.User.FirstName
		}
		if e.User.LastName != "" {
			claims["family_name"] = e.User.LastName
		}
		if e.User.Username != "" {
			claims["preferred_username"] = e.User.Username
		}
		if e.User.AvatarURL != "" {
			claims["picture"] = e.User.AvatarURL
		}
	}

	if ScopesContain(scope, "email") {
		email := e.EmailAuth.Email
		if email == "" && e.User != nil {
			email = e.User.Email
		}
		if email != "" {
			claims["email"] = email
			// The email auth record is the credential the user signs in with,
			// so we treat it as verified.
			claims["email_verified"] = true
		}
	}

	return claims
}

// BuildIDTokenClaims assembles the OIDC-specific custom claims for an ID token.
// Registered claims (iss/sub/aud/exp/iat/jti) are stamped by core at signing
// time; this only supplies the identity, auth_time, nonce, and at_hash claims.
func BuildIDTokenClaims(entityID string, scope string, nonce string, accessToken string, authTime int64) map[string]interface{} {
	e, _ := fetchOIDCEntity(entityID)
	claims := identityClaims(e, scope)
	claims["auth_time"] = authTime
	if nonce != "" {
		claims["nonce"] = nonce
	}
	if accessToken != "" {
		claims["at_hash"] = AccessTokenHash(accessToken)
	}
	return claims
}

// BuildUserInfoClaims returns the UserInfo response for an entity, filtered by
// the access token's granted scopes. `sub` is always present per spec.
func BuildUserInfoClaims(entityID string, scope string) (map[string]interface{}, error) {
	e, err := fetchOIDCEntity(entityID)
	if err != nil {
		return nil, err
	}
	claims := identityClaims(e, scope)
	claims["sub"] = entityID
	return claims, nil
}

// AccessTokenHash computes the OIDC `at_hash`: the base64url-encoded left-most
// half of the SHA-256 of the access token. RS256 uses SHA-256, so that's the
// left 128 bits (16 bytes).
func AccessTokenHash(accessToken string) string {
	sum := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}
