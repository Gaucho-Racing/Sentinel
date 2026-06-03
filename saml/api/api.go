package api

import (
	"time"

	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
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
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET("/saml/ping", Ping)

	// IdP endpoints served at the issuer root (gateway routes these without
	// stripping a prefix), so the URLs published in metadata are clean.
	router.GET("/saml/metadata", Metadata)
	router.GET("/saml/sso", SSO)
	router.POST("/saml/sso", SSO)

	// Consent endpoints reached through the gateway's /api prefix (stripped to
	// /saml/authorize). The SPA holds the first-party session and drives these.
	router.GET("/saml/authorize", ValidateAuthorize)
	router.POST("/saml/authorize", Authorize)
}

// GetClientIP returns the originating client IP, preferring Cloudflare's
// unspoofable CF-Connecting-IP and falling back to gin's c.ClientIP().
func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
