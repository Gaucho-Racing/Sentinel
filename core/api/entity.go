package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMe(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "user:read"))
	id := GetRequestTokenEntityID(c)

	entity, err := service.GetEntityByID(id)
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

func GetEntity(c *gin.Context) {
	id := c.Param("id")
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "user:read") && RequestTokenHasEntityID(c, id),
	))

	entity, err := service.GetEntityByID(id)
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

func GetEntityByID(c *gin.Context) {
	entityID := c.Param("entityID")
	entity, err := service.GetEntityByID(entityID)
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

func GetEntityGroups(c *gin.Context) {
	entityID := c.Param("entityID")
	groups, err := service.GetGroupsForEntity(entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

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
	clientID := c.Query("client_id")
	scope := c.Query("scope")
	limit := c.Query("limit")

	logins, err := service.GetEntityLogins(entityID, clientID, scope, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logins)
}
