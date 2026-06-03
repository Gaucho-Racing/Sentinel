package api

import (
	"net/http"
	"strings"

	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

// UserInfo is the OIDC UserInfo endpoint. It authenticates with the access
// token issued during the flow and returns the standard claims the token's
// scopes allow. Requires the openid scope.
func UserInfo(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.Header("WWW-Authenticate", "Bearer")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	var claims map[string]interface{}
	if err := sentinel.Post("/api/core/token/validate", map[string]string{"token": token}, &claims); err != nil {
		c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	entityID, _ := claims["sub"].(string)
	scope, _ := claims["scope"].(string)
	if entityID == "" || !service.ScopesContain(scope, "openid") {
		c.JSON(http.StatusForbidden, gin.H{"error": "token lacks the openid scope"})
		return
	}

	// aud is the client the token was issued to — used to apply the same
	// per-client group filtering the access token gets. Per RFC 7519 it may be
	// serialized as either an array or a single string.
	clientID := ""
	switch aud := claims["aud"].(type) {
	case []interface{}:
		if len(aud) > 0 {
			clientID, _ = aud[0].(string)
		}
	case string:
		clientID = aud
	}

	info, err := service.BuildUserInfoClaims(entityID, clientID, scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user info"})
		return
	}

	c.JSON(http.StatusOK, info)
}
