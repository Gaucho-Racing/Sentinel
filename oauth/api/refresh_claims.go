package api

import (
	"errors"

	"github.com/gaucho-racing/sentinel/oauth/service"
)

var errInvalidRefreshTokenClaims = errors.New("invalid refresh token claims")

type refreshTokenClaims struct {
	EntityID string
	Scope    string
	TokenID  string
}

func parseRefreshTokenClaims(claims map[string]interface{}, expectedAudience string) (refreshTokenClaims, error) {
	entityID, _ := claims["sub"].(string)
	scope, _ := claims["scope"].(string)
	tokenID, _ := claims["jti"].(string)
	if entityID == "" || tokenID == "" || !service.ScopesContain(scope, "refresh_token") {
		return refreshTokenClaims{}, errInvalidRefreshTokenClaims
	}
	if !audienceMatches(claims["aud"], expectedAudience) {
		return refreshTokenClaims{}, errInvalidRefreshTokenClaims
	}
	return refreshTokenClaims{EntityID: entityID, Scope: scope, TokenID: tokenID}, nil
}

func audienceMatches(raw interface{}, expected string) bool {
	if expected == "" {
		return false
	}
	switch audience := raw.(type) {
	case string:
		return audience == expected
	case []interface{}:
		if len(audience) != 1 {
			return false
		}
		value, ok := audience[0].(string)
		return ok && value == expected
	case []string:
		return len(audience) == 1 && audience[0] == expected
	default:
		return false
	}
}
