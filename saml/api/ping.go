package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": config.FormattedNameWithVersion() + " is online!"})
}
