package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// allowedSATTLs accepts the four TTL choices the UI surfaces — every
// other value gets rejected with a 400 so callers can't bypass the
// dropdown to mint a 17-day key or similar oddity. 0 means "never expires"
// (a ~100-year exp on the JWT under the hood).
var allowedSATTLs = map[int]struct{}{
	30: {}, 90: {}, 365: {}, 0: {},
}

// requireAppOwnerOrAdmin gates write access to an app's service accounts.
// Mirrors the existing pattern on ApplicationDetailsPage: the app's owner
// can manage their own resources, and global admins can override. Returns
// the resolved app on success — the caller usually needs it anyway, so
// returning it here saves a second lookup.
func requireAppOwnerOrAdmin(c *gin.Context, appID string) (model.Application, bool) {
	app, err := service.GetApplicationByID(appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return model.Application{}, false
	}
	if !Any(
		RequestTokenHasEntityID(c, app.OwnerID),
		RequestUserIsAdmin(c),
	) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you are not authorized to manage this application"})
		return model.Application{}, false
	}
	return app, true
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
	// Scope must be a space-separated subset of
	// service.ServiceAccountAllowedScopes. Empty string is allowed.
	Scope string `json:"scope"`
	// TTLDays must be one of {30, 90, 365, 0}. 0 = never expires
	// (issued as a ~100-year JWT under the hood).
	TTLDays int `json:"ttl_days"`
}

type serviceAccountWithToken struct {
	ServiceAccount model.ServiceAccount `json:"service_account"`
	// Token is the raw signed JWT. Surfaced ONCE on create + rotate;
	// subsequent reads of the SA only see the row metadata via
	// ActiveToken (no raw string).
	Token string `json:"token"`
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
	if err := service.ValidateServiceAccountScope(req.Scope); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, ok := allowedSATTLs[req.TTLDays]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ttl_days must be one of 30, 90, 365, or 0 (never)"})
		return
	}

	sa, err := service.CreateServiceAccountForApp(id, name, req.Scope, req.TTLDays, GetRequestTokenEntityID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, raw, err := service.MintServiceAccountToken(sa)
	if err != nil {
		// Best-effort rollback so a failed mint doesn't leave an SA
		// with no token — easier to delete-and-retry than to debug an
		// orphan row.
		_ = service.DeleteServiceAccount(sa.ID)
		_ = service.DeleteEntity(sa.EntityID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Re-populate so ActiveToken is set on the returned SA.
	service.PopulateServiceAccount(&sa)
	logger.SugarLogger.Infof("Created service account %s (entity=%s) for application %s", sa.ID, sa.EntityID, id)
	c.JSON(http.StatusOK, serviceAccountWithToken{ServiceAccount: sa, Token: raw})
}

// RotateServiceAccountToken delete-and-mints a fresh JWT for the SA,
// reusing the SA's stored scope + ttl_days. The previous token is
// revoked at the DB layer immediately — any client still using it
// starts getting 401s. Returns the new raw JWT (shown once).
func RotateServiceAccountToken(c *gin.Context) {
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

	_, raw, err := service.MintServiceAccountToken(sa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	service.PopulateServiceAccount(&sa)
	logger.SugarLogger.Infof("Rotated token for service account %s", sa.ID)
	c.JSON(http.StatusOK, serviceAccountWithToken{ServiceAccount: sa, Token: raw})
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

	// Revoke the SA's token before deleting the SA itself so an
	// outstanding token can't briefly authenticate against a now-
	// orphaned SA. DeleteTokensForEntity is idempotent — no error if
	// the SA had no active token.
	if err := service.DeleteTokensForEntity(sa.EntityID); err != nil {
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
