package api

import (
	"net/http"
	"sort"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/model"
	"github.com/gin-gonic/gin"
)

// OpenIDConfiguration serves the OIDC discovery document. Endpoint URLs are
// derived from the configured issuer (the public base URL). The browser-facing
// authorization endpoint is the SPA consent route (no /api prefix); the
// token/userinfo endpoints are backend routes behind the gateway's /api prefix;
// the JWKS lives on core.
func OpenIDConfiguration(c *gin.Context) {
	issuer := config.Issuer
	c.JSON(http.StatusOK, gin.H{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/api/oauth/token",
		"userinfo_endpoint":                     issuer + "/api/oauth/userinfo",
		"jwks_uri":                              issuer + "/api/core/keys",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
		"scopes_supported":                      supportedScopes(),
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat", "jti", "auth_time", "nonce", "at_hash",
			"name", "given_name", "family_name", "preferred_username", "picture",
			"email", "email_verified",
		},
	})
}

// supportedScopes lists the scopes a third-party client may request — every
// valid scope except the reserved first-party sentinel:all.
func supportedScopes() []string {
	scopes := make([]string, 0, len(model.ValidScopes))
	for scope := range model.ValidScopes {
		if scope == "sentinel:all" {
			continue
		}
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes
}
