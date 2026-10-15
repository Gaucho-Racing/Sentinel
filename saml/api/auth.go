package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

// AuthChecker is a soft middleware: when Authorization: Bearer is
// present it validates the JWT against core's /core/token/validate and
// stashes (sub, scope) on the context. Handlers that need auth call
// Require(...) themselves; the public endpoints (ping, metadata,
// /saml/sso) keep working without a bearer.
//
// Mirrors core/api/AuthChecker so handlers can use the same Require /
// RequestTokenHas* helpers core does.
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
			if sub, ok := claims["sub"].(string); ok {
				c.Set("Auth-EntityID", sub)
			}
			if scope, ok := claims["scope"].(string); ok {
				c.Set("Auth-Scope", scope)
			}
			c.Set("Auth-Claims", claims)
		}
		c.Next()
	}
}

// UnauthorizedPanicHandler converts Require()'s panic into a 401. Any
// other panic is logged and returned as a 500. Same shape as core.
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

func GetRequestTokenEntityID(c *gin.Context) string {
	id, ok := c.Get("Auth-EntityID")
	if !ok {
		return ""
	}
	return id.(string)
}

func GetRequestToken(c *gin.Context) string {
	token, _ := c.Get("Auth-Token")
	value, _ := token.(string)
	return value
}

func RequestTokenHasEntityID(c *gin.Context, entityID string) bool {
	return GetRequestTokenEntityID(c) == entityID
}

// RequestTokenHasInternalAccess mirrors core's check for Sentinel's own
// service accounts: sentinel:internal on the sentinel audience.
func RequestTokenHasInternalAccess(c *gin.Context) bool {
	value, _ := c.Get("Auth-Claims")
	claims, _ := value.(map[string]interface{})
	if claims == nil {
		return false
	}
	scope, _ := claims["scope"].(string)
	if !slices.Contains(strings.Fields(scope), "sentinel:internal") {
		return false
	}
	audienceOK := false
	switch audience := claims["aud"].(type) {
	case string:
		audienceOK = audience == "sentinel"
	case []interface{}:
		for _, value := range audience {
			if value == "sentinel" {
				audienceOK = true
			}
		}
	}
	accountType, _ := claims["type"].(string)
	accountID, _ := claims["service_account_id"].(string)
	return audienceOK && accountType == "service_account" && accountID != ""
}
