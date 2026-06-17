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
	// Entity rows carry PII (email-auth, phone-auth, linked external
	// identities, user profile). Self can read their own; admin and
	// internal automation override.
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasEntityID(c, entityID),
		RequestUserIsAdmin(c),
	))
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
	// Group membership is an authorization signal — leaking another
	// user's groups would tell an attacker who has admin-equivalent
	// access. Self / admin / internal only.
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasEntityID(c, entityID),
		RequestUserIsAdmin(c),
	))
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
	// Raw GroupMember rows (with source labels) are used by integration
	// services to diff their own writes — same self/admin/internal
	// trust level as GetEntityGroups.
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasEntityID(c, entityID),
		RequestUserIsAdmin(c),
	))
	source := c.Query("source")
	memberships, err := service.GetMembershipsForEntity(entityID, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, memberships)
}

func GetEntityByExternalAuth(c *gin.Context) {
	// Reverse-lookup of "which Sentinel entity does Discord user X
	// map to?" — leaks the user/Discord identity pairing. Reserved for
	// internal automation; the oauth-discord-login flow is the
	// canonical caller.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
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
	// Enumeration of every onboarded user for a provider — used by
	// the discord sync's full sweep. Internal callers only.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
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
	entityID := c.Param("entityID")
	// Login history is an audit-grade signal. Self / admin / internal
	// only.
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasEntityID(c, entityID),
		RequestUserIsAdmin(c),
	))
	logins, err := service.GetEntityLogins(service.EntityLoginsFilter{
		EntityID: entityID,
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
