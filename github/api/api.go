package api

import (
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/github/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router(server *service.Server) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:          12 * time.Hour,
	}))
	r.GET("/github/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/github/link", server.BeginLink)
	r.DELETE("/github/link", server.Unlink)
	r.GET("/github/link/status", server.LinkStatus)
	r.GET("/github/callback", server.FinishLink)
	return r
}
