package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMe(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "user:read"),
	))
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
		RequestTokenHasScope(c, "sentinel:all"),
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

// GetEntityMemberships returns the raw GroupMember rows for an entity,
// optionally filtered by source via the ?source= query param. Used by
// integration services to read their own membership writes for diffing.
func GetEntityMemberships(c *gin.Context) {
	entityID := c.Param("entityID")
	source := c.Query("source")
	memberships, err := service.GetMembershipsForEntity(entityID, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, memberships)
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

func ListExternalAuthsByProvider(c *gin.Context) {
	provider := c.Param("provider")
	auths, err := service.ListExternalAuthsByProvider(provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auths)
}

func CreateEntityLogin(c *gin.Context) {
	// Login rows are the audit log of token issuance. Writing into
	// them is reserved for the oauth service (which records each
	// session it mints); admins/users shouldn't be backdating their
	// own entries.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
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
	logins, err := service.GetEntityLogins(service.EntityLoginsFilter{
		EntityID: c.Param("entityID"),
		ClientID: c.Query("client_id"),
		Scope:    c.Query("scope"),
		Before:   c.Query("before"),
		After:    c.Query("after"),
		Limit:    c.Query("limit"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logins)
}
