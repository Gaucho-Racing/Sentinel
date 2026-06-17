package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllApplications(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:read"),
	))
	applications, err := service.GetAllApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applications)
}

func GetApplicationByID(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:read"),
	))
	id := c.Param("id")
	app, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

func GetApplicationByClientID(c *gin.Context) {
	// Resolving an app from a client_id leaks the owner_id and
	// internal metadata. The oauth/saml services use it to look up
	// the app a token request is targeting, so internal automation
	// must work; admins also have full read.
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestUserIsAdmin(c),
	))
	clientID := c.Param("clientID")
	app, err := service.GetApplicationByClientID(clientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

type verifyClientRequest struct {
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

func VerifyClientCredentials(c *gin.Context) {
	var req verifyClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := service.GetApplicationByClientID(req.ClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}
	if app.ClientSecret != req.ClientSecret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "valid"})
}

type createApplicationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	LaunchURL   string `json:"launch_url"`
}

// createdApplicationResponse exposes the freshly minted client_secret
// (model.Application JSON-skips it on subsequent reads).
type createdApplicationResponse struct {
	model.Application
	Secret string `json:"client_secret"`
}

func CreateApplication(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:write"),
	))

	var req createApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := service.CreateApplication(model.Application{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		LaunchURL:   req.LaunchURL,
		OwnerID:     GetRequestTokenEntityID(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createdApplicationResponse{
		Application: app,
		Secret:      app.ClientSecret,
	})
}

type updateApplicationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	LaunchURL   string `json:"launch_url"`
}

func UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	var req updateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.Name = req.Name
	existing.Description = req.Description
	existing.IconURL = req.IconURL
	existing.LaunchURL = req.LaunchURL
	updated, err := service.UpdateApplication(existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// GetApplicationSecret returns just the client_secret. Separated from the
// main GET handler so the secret doesn't leak through list/by-id reads.
// Gating tightens later when ownership lands; for now any first-party
// bearer can see it.
func GetApplicationSecret(c *gin.Context) {
	id := c.Param("id")
	app, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasAudience(c, "sentinel") && RequestTokenHasEntityID(c, app.OwnerID),
	))
	c.JSON(http.StatusOK, gin.H{"client_secret": app.ClientSecret})
}

func DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	if err := service.DeleteApplication(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "application deleted"})
}

func GetApplicationGroups(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:read"),
	))
	id := c.Param("id")
	groups, err := service.GetGroupsForApplication(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetApplicationGroupsByClientID returns an application's group links resolved
// by client_id. Internal-only route — the oauth service uses it to resolve the
// groups claim and enforce the access gate. Now that oauth carries its own SA
// bearer, the gate is sentinel:all (matches the other internal-only reads).
func GetApplicationGroupsByClientID(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	clientID := c.Param("clientID")
	app, err := service.GetApplicationByClientID(clientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	groups, err := service.GetGroupsForApplication(app.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

type upsertApplicationGroupRequest struct {
	GroupID  string `json:"group_id" binding:"required"`
	Required bool   `json:"required"`
}

// AddApplicationGroup upserts the (application, group) link. If the link
// already exists, the Required flag is updated in place — no PATCH needed.
func AddApplicationGroup(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	var req upsertApplicationGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ag, err := service.UpsertApplicationGroup(model.ApplicationGroup{
		ApplicationID: id,
		GroupID:       req.GroupID,
		Required:      req.Required,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ag)
}

func RemoveApplicationGroup(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	groupID := c.Param("groupID")
	if err := service.DeleteApplicationGroup(id, groupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group removed from application"})
}

func GetApplicationRedirectURIs(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:read"),
	))
	id := c.Param("id")
	uris, err := service.GetRedirectURIsForApplication(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, uris)
}

type addRedirectURIRequest struct {
	RedirectURI string `json:"redirect_uri" binding:"required"`
}

func AddApplicationRedirectURI(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	var req addRedirectURIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uri, err := service.CreateApplicationRedirectURI(id, req.RedirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, uri)
}

func RemoveApplicationRedirectURI(c *gin.Context) {
	id := c.Param("id")
	existing, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, existing))
	uri := c.Query("uri")
	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uri query parameter is required"})
		return
	}
	if err := service.DeleteApplicationRedirectURI(id, uri); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "redirect uri removed from application"})
}

// ApplicationWriteAuthorized returns true when the bearer can mutate the
// given application: admin scope (sentinel:all), first-party UI used by the
// owner OR an Admins-group member, or a third-party token with
// applications:write granted by the owner.
func ApplicationWriteAuthorized(c *gin.Context, app model.Application) bool {
	return Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasAudience(c, "sentinel") && RequestTokenHasEntityID(c, app.OwnerID),
		RequestTokenHasAudience(c, "sentinel") && RequestUserIsAdmin(c),
		RequestTokenHasScope(c, "applications:write") && RequestTokenHasEntityID(c, app.OwnerID),
	)
}
