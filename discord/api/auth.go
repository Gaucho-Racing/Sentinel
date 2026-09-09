package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gaucho-racing/sentinel/discord/authz"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

// AuthChecker is a soft middleware: if Authorization: Bearer is present
// it asks core to validate the JWT and stashes the resulting claims on
// the context. Handlers that require auth call Require(...) themselves
// — endpoints that don't (ping, onboarding-token reads protected by
// token-as-secret) keep working without a bearer.
//
// Mirrors core/api/AuthChecker so handlers in this service can use the
// same Require/Any/RequestTokenHasScope helpers core does.
func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			var claims map[string]interface{}
			if err := sentinel.Post("/api/core/token/validate", map[string]string{"token": token}, &claims); err != nil {
				logger.SugarLogger.Errorf("Failed to validate token: %v", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.Set("Auth-Token", token)
			c.Set("Auth-Claims", claims)
			if sub, ok := claims["sub"].(string); ok {
				c.Set("Auth-EntityID", sub)
			}
			if scope, ok := claims["scope"].(string); ok {
				c.Set("Auth-Scope", scope)
			}
		}
		c.Next()
	}
}

// UnauthorizedPanicHandler converts Require()'s panic into a 403. Any
// other panic is logged and rethrown as a 500. Same shape as core.
func UnauthorizedPanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if err == "Unauthorized" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized to access this resource"})
					return
				}
				logger.SugarLogger.Errorf("Unexpected panic: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
		}()
		c.Next()
	}
}

func Require(c *gin.Context, condition bool) {
	if !condition {
		panic("Unauthorized")
	}
}

func RequestTokenHasScope(c *gin.Context, scope string) bool {
	scopes, ok := c.Get("Auth-Scope")
	if !ok {
		return false
	}
	return authz.HasScope(scopes.(string), scope)
}

func GetRequestToken(c *gin.Context) string {
	token, _ := c.Get("Auth-Token")
	value, _ := token.(string)
	return value
}

func GetRequestTokenClaims(c *gin.Context) map[string]any {
	claims, _ := c.Get("Auth-Claims")
	value, _ := claims.(map[string]any)
	return value
}

func RequestTokenHasInternalAccess(c *gin.Context) bool {
	claims := GetRequestTokenClaims(c)
	return authz.IsInternalServiceAccount(
		getRequestTokenScopes(c),
		claims["aud"],
		claims,
	)
}

func RequestTokenHasFirstPartyAccess(c *gin.Context) bool {
	claims := GetRequestTokenClaims(c)
	return authz.IsFirstPartyUser(getRequestTokenScopes(c), claims["aud"], claims)
}

func RequestTokenCanManageGroup(c *gin.Context, groupID string) bool {
	if RequestTokenHasInternalAccess(c) {
		return true
	}
	if !RequestTokenHasFirstPartyAccess(c) && !RequestTokenHasScope(c, authz.GroupsWriteScope) {
		return false
	}
	return requestCoreAccessCheck(c, "/api/groups/"+url.PathEscape(groupID)+"/write-access")
}

func RequestTokenHasAdminAccess(c *gin.Context) bool {
	if RequestTokenHasInternalAccess(c) {
		return true
	}
	return RequestTokenHasFirstPartyAccess(c) && requestCoreAccessCheck(c, "/api/entities/@me/admin-access")
}

func getRequestTokenScopes(c *gin.Context) string {
	scopes, _ := c.Get("Auth-Scope")
	value, _ := scopes.(string)
	return value
}

func requestCoreAccessCheck(c *gin.Context, route string) bool {
	token := GetRequestToken(c)
	if token == "" {
		return false
	}
	return sentinel.Get(route, nil, map[string]string{"Authorization": "Bearer " + token}) == nil
}
