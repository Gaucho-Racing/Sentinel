package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// requireSAManagable looks up the SA, then defers to requireAppOwnerOrAdmin
// against the app the SA belongs to. Returns the resolved SA on success
// since the caller almost always needs it.
func requireSAManagable(c *gin.Context, saID string) (model.ServiceAccount, bool) {
	sa, err := service.GetServiceAccountByID(saID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return model.ServiceAccount{}, false
	}
	if _, ok := requireAppOwnerOrAdmin(c, sa.ApplicationID); !ok {
		return model.ServiceAccount{}, false
	}
	return sa, true
}

func ListAPIKeysForServiceAccount(c *gin.Context) {
	saID := c.Param("id")
	if _, ok := requireSAManagable(c, saID); !ok {
		return
	}
	keys, err := service.ListAPIKeysForServiceAccount(saID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

type createAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
	// TTLDays: 0 means never expires. Otherwise interpreted as days from
	// now. The web UI presents 30/90/365/never; the server only cares
	// about the numeric value.
	TTLDays int `json:"ttl_days"`
	// Scope is a space-separated list following the OAuth convention.
	// Validation that this is a subset of the SA's allowed scopes happens
	// at use time (the auth middleware sets Auth-Scope from this value;
	// per-endpoint scope checks decide if it's enough). Empty string is
	// allowed — that key has no scope and can only be used for endpoints
	// that don't require any.
	Scope string `json:"scope"`
}

type createAPIKeyResponse struct {
	Key model.APIKey `json:"key"`
	// Token is the raw sk_..._... string. Surfaced ONCE on creation;
	// subsequent reads (GET /api-keys) never include it.
	Token string `json:"token"`
}

func CreateAPIKeyForServiceAccount(c *gin.Context) {
	saID := c.Param("id")
	if _, ok := requireSAManagable(c, saID); !ok {
		return
	}
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.TTLDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ttl_days cannot be negative"})
		return
	}

	var expiresAt *time.Time
	if req.TTLDays > 0 {
		t := time.Now().Add(time.Duration(req.TTLDays) * 24 * time.Hour)
		expiresAt = &t
	}

	key, raw, err := service.GenerateAPIKey(saID, name, req.Scope, expiresAt, GetRequestTokenEntityID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.SugarLogger.Infof("Created API key %s for service account %s", key.ID, saID)
	c.JSON(http.StatusOK, createAPIKeyResponse{Key: key, Token: raw})
}

func RevokeAPIKey(c *gin.Context) {
	saID := c.Param("id")
	keyID := c.Param("keyID")
	if _, ok := requireSAManagable(c, saID); !ok {
		return
	}
	if err := service.DeleteAPIKey(keyID, saID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "api key revoked"})
}
