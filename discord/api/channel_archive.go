package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
	"github.com/gin-gonic/gin"
)

func GetArchivedChannels(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	records, err := service.GetAllArchivedChannels()
	if err != nil {
		logger.SugarLogger.Errorf("Failed to fetch archived channels: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch archived channels"})
		return
	}
	c.JSON(http.StatusOK, records)
}
