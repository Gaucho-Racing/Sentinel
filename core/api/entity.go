package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetEntityByExternalAuth(c *gin.Context) {
	provider := c.Param("provider")
	externalID := c.Param("externalID")
	entity, err := service.GetEntityByExternalAuth(provider, externalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entity)
}
