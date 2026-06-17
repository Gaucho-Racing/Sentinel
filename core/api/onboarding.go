package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createEntityRequest struct {
	Type model.EntityType `json:"type" binding:"required"`
}

func CreateEntity(c *gin.Context) {
	// Entities are the root identity row that everything else hangs
	// off (users, service accounts, group memberships). Creation is
	// reserved for internal automation — the discord onboarding flow
	// is the canonical caller and now carries sentinel:all via its SA.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	var req createEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := service.CreateEntity(model.Entity{Type: req.Type})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entity)
}

type createEmailAuthRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateEntityEmailAuth upserts an entity's email auth — creates a fresh
// row on onboarding, replaces email + password on subsequent calls
// (password reset, email change). Picks the service-layer function based
// on whether a row already exists.
func CreateEntityEmailAuth(c *gin.Context) {
	// This endpoint upserts an entity's email + password — the entire
	// account-takeover primitive in one call. Reserved for internal
	// callers (discord onboarding mints the initial email auth; future
	// password-reset flows would go through a separate token-mediated
	// path before hitting this handler).
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	entityID := c.Param("entityID")
	var req createEmailAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existing, err := service.GetEmailAuthForEntity(entityID)
	var auth model.EntityEmail
	switch {
	case err == gorm.ErrRecordNotFound || existing.EntityID == "":
		auth, err = service.CreateEmailAuthForEntity(entityID, req.Email, hashed)
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	default:
		auth, err = service.UpdateEmailAuthForEntity(entityID, req.Email, hashed)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auth)
}

type createPhoneAuthRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func CreateEntityPhoneAuth(c *gin.Context) {
	// Same trust level as the other entity-auth writers: internal
	// onboarding only.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	entityID := c.Param("entityID")
	var req createPhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	auth, err := service.CreatePhoneAuthForEntity(entityID, req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auth)
}

type createExternalAuthRequest struct {
	Provider   model.ExternalAuthProvider `json:"provider" binding:"required"`
	ExternalID string                     `json:"external_id" binding:"required"`
}

func CreateEntityExternalAuth(c *gin.Context) {
	// Linking an external identity (DISCORD, GITHUB, etc.) to an
	// entity is account-takeover-adjacent — anyone able to write this
	// row can claim any entity. Internal callers only.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	entityID := c.Param("entityID")
	var req createExternalAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	auth, err := service.CreateExternalAuthForEntity(model.EntityExternalAuth{
		EntityID:   entityID,
		Provider:   req.Provider,
		ExternalID: req.ExternalID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auth)
}

type updateExternalAuthMetadataRequest struct {
	Metadata model.JSONMap `json:"metadata" binding:"required"`
}

// UpdateEntityExternalAuthMetadata refreshes the per-provider metadata jsonb
// on an existing external auth row. Login handlers call this on every
// successful provider sign-in so the email / username / avatar that came
// back from the provider stays current.
func UpdateEntityExternalAuthMetadata(c *gin.Context) {
	// Called by login handlers on every successful provider sign-in
	// (oauth-discord-login refreshes the cached email/username/avatar
	// after a successful Discord exchange). Internal callers only.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	entityID := c.Param("entityID")
	provider := c.Param("provider")
	var req updateExternalAuthMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.UpdateExternalAuthMetadata(entityID, provider, req.Metadata); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "external auth not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
