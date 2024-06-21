package controller

import (
	"net/http"
	"sentinel/config"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Sentinel v" + config.Version + " is online!"})
}
