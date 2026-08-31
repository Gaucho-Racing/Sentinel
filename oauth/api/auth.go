package api

import (
	"net/http"
	"strings"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		var claims map[string]interface{}
		if err := sentinel.Post("/api/core/token/validate", map[string]string{"token": token}, &claims); err != nil {
			logger.SugarLogger.Errorf("Failed to validate token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		entityID, _ := claims["sub"].(string)
		if entityID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token subject is required"})
			return
		}

		c.Set("Auth-Token", token)
		c.Set("Auth-EntityID", entityID)
		c.Next()
	}
}

func UnauthorizedPanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if recovered == "Unauthorized" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized to access this resource"})
					return
				}
				logger.SugarLogger.Errorf("Unexpected panic: %v", recovered)
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
	entityID, _ := c.Get("Auth-EntityID")
	value, _ := entityID.(string)
	return value
}
