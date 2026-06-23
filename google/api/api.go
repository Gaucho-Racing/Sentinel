package api

import (
	"time"

	"github.com/gaucho-racing/sentinel/google/config"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
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
	router.GET("/google/ping", Ping)

	router.GET("/google/group-bindings", ListGoogleBindings)
	router.POST("/google/group-bindings", CreateGoogleBinding)
	router.DELETE("/google/group-bindings/:bindingID", DeleteGoogleBinding)

	router.POST("/google/reconcile", TriggerReconcile)
}

// GetClientIP returns the originating client IP, preferring Cloudflare's
// unspoofable CF-Connecting-IP and falling back to gin's c.ClientIP().
func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
