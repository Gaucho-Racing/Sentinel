package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// requireAppOwnerOrAdmin gates write access to an app's service accounts.
// Mirrors the existing pattern on ApplicationDetailsPage: the app's owner
// can manage their own resources, and global admins can override. Returns
// (app, true) on success — the caller usually needs the app anyway, so
// returning it here saves a second lookup.
func requireAppOwnerOrAdmin(c *gin.Context, appID string) (string, bool) {
	app, err := service.GetApplicationByID(appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return "", false
	}
	if !Any(
		RequestTokenHasEntityID(c, app.OwnerID),
		RequestUserIsAdmin(c),
	) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you are not authorized to manage this application"})
		return "", false
	}
	return app.ID, true
}

func ListServiceAccountsForApplication(c *gin.Context) {
	id := c.Param("id")
	if _, ok := requireAppOwnerOrAdmin(c, id); !ok {
		return
	}
	sas, err := service.GetServiceAccountsByApplicationID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sas)
}

type createServiceAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

func CreateServiceAccountForApp(c *gin.Context) {
	id := c.Param("id")
	if _, ok := requireAppOwnerOrAdmin(c, id); !ok {
		return
	}
	var req createServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	sa, err := service.CreateServiceAccountForApp(id, name, GetRequestTokenEntityID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.SugarLogger.Infof("Created service account %s (entity=%s) for application %s", sa.ID, sa.EntityID, id)
	c.JSON(http.StatusOK, sa)
}

func DeleteServiceAccount(c *gin.Context) {
	id := c.Param("id")
	sa, err := service.GetServiceAccountByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, ok := requireAppOwnerOrAdmin(c, sa.ApplicationID); !ok {
		return
	}

	// Revoke all keys before deleting the SA itself so an outstanding key
	// can't briefly authenticate against a now-orphaned SA (the SA's
	// Entity row still exists until we delete it below; an in-flight
	// request could otherwise see a brief window of valid-key /
	// missing-SA).
	if err := service.DeleteAllAPIKeysForServiceAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := service.DeleteServiceAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Best-effort entity cleanup. A leftover SERVICE_ACCOUNT entity is
	// harmless (no auth path resolves it), so we log on failure rather
	// than fail the request.
	if err := service.DeleteEntity(sa.EntityID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.SugarLogger.Errorf("Failed to delete entity %s for SA %s: %v", sa.EntityID, id, err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "service account deleted"})
}
