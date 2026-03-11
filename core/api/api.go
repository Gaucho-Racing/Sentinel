package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	api := InitializeRouter()
	InitializeRoutes(api)
	err := api.Run(":" + config.Port)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to start server: %v", err)
	}
}

func InitializeRouter() *gin.Engine {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))
	r.Use(AuthChecker())
	r.Use(UnauthorizedPanicHandler())
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET(fmt.Sprintf("/%s/ping", config.Service.Name), Ping)
	router.GET(fmt.Sprintf("/%s/keys", config.Service.Name), JWKS)
}

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				claims, err := service.ValidateAccessToken(strings.Split(authHeader, "Bearer ")[1])
				if err != nil {
					logger.SugarLogger.Errorln("Failed to validate token: " + err.Error())
					c.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
				} else if strings.Contains(claims.Scope, "refresh_token") {
					logger.SugarLogger.Errorln("Received refresh token instead of access token")
					c.AbortWithStatusJSON(401, gin.H{"error": "Received refresh token instead of access token"})
				} else {
					logger.SugarLogger.Infof("Decoded token: %s (%s)", claims.ID, claims.Subject)
					logger.SugarLogger.Infof("↳ Client ID: %s", claims.Audience[0])
					logger.SugarLogger.Infof("↳ Scope: %s", claims.Scope)
					logger.SugarLogger.Infof("↳ Issued at: %s", claims.IssuedAt.String())
					logger.SugarLogger.Infof("↳ Expires at: %s", claims.ExpiresAt.String())
					c.Set("Auth-Token", strings.Split(authHeader, "Bearer ")[1])
					c.Set("Auth-EntityID", claims.Subject)
					c.Set("Auth-Audience", claims.Audience[0])
					c.Set("Auth-Scope", claims.Scope)
				}
			}
		}
		c.Next()
	}
}

func UnauthorizedPanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if err == "Unauthorized" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized to access this resource"})
				} else {
					// Handle other panics
					logger.SugarLogger.Errorf("Unexpected panic: %v", err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.(string)})
				}
			}
		}()
		c.Next()
	}
}

// Require checks if a condition is true, otherwise aborts the request
func Require(c *gin.Context, condition bool) {
	if !condition {
		panic("Unauthorized")
	}
}

// Any checks if any condition is true, otherwise returns false
func Any(conditions ...bool) bool {
	for _, condition := range conditions {
		if condition {
			return true
		}
	}
	return false
}

// All checks if all conditions are true, otherwise returns false
func All(conditions ...bool) bool {
	for _, condition := range conditions {
		if !condition {
			return false
		}
	}
	return true
}
