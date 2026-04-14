package api

import (
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/core/model"
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

func CreateEntityLogin(c *gin.Context) {
	var login model.EntityLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	login, err := service.CreateEntityLogin(login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, login)
}

func GetEntityLogins(c *gin.Context) {
	entityID := c.Param("entityID")
	logins, err := service.GetEntityLoginsByEntityID(entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logins)
}

func CheckRecentLogin(c *gin.Context) {
	entityID := c.Query("entity_id")
	clientID := c.Query("client_id")
	scope := c.Query("scope")
	if entityID == "" || clientID == "" || scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id, client_id, and scope are required"})
		return
	}
	recent := service.HasRecentLogin(entityID, clientID, scope, 7*24*time.Hour)
	c.JSON(http.StatusOK, gin.H{"recent": recent})
}
