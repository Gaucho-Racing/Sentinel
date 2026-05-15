package api

import (
	"time"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
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
	router.GET("/oauth/ping", Ping)
	router.GET("/oauth/authorize", ValidateAuthorize)
	router.POST("/oauth/authorize", Authorize)
	router.POST("/oauth/token", ExchangeToken)

	router.POST("/auth/login/email-password", LoginEmailPassword)
	router.POST("/auth/refresh", RefreshSession)
}

// GetClientIP returns the originating client IP. Prefers Cloudflare's
// CF-Connecting-IP (set by CF, overrides any client-supplied value, and
// unspoofable as long as the origin only accepts traffic from CF's IP
// ranges). Falls back to gin's c.ClientIP() in dev or non-CF deployments.
func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
