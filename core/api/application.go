package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllApplications(c *gin.Context) {
	applications, err := service.GetAllApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applications)
}

func GetApplicationByID(c *gin.Context) {
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
	Require(c, RequestTokenExists(c))
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
	Require(c, RequestTokenHasAudience(c, "sentinel"))
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
	c.JSON(http.StatusOK, gin.H{"client_secret": app.ClientSecret})
}

func DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteApplication(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "application deleted"})
}

func GetApplicationGroups(c *gin.Context) {
	id := c.Param("id")
	groups, err := service.GetGroupsForApplication(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

type addApplicationGroupRequest struct {
	GroupID string `json:"group_id" binding:"required"`
}

func AddApplicationGroup(c *gin.Context) {
	id := c.Param("id")
	var req addApplicationGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ag, err := service.CreateApplicationGroup(model.ApplicationGroup{
		ApplicationID: id,
		GroupID:       req.GroupID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ag)
}

func RemoveApplicationGroup(c *gin.Context) {
	id := c.Param("id")
	groupID := c.Param("groupID")
	if err := service.DeleteApplicationGroup(id, groupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group removed from application"})
}

func GetApplicationRedirectURIs(c *gin.Context) {
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
