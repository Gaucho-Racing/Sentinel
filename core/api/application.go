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

func CreateOrUpdateApplication(c *gin.Context) {
	var app model.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, err := service.GetApplicationByID(app.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.ID != "" {
		app, err = service.UpdateApplication(app)
	} else {
		app, err = service.CreateApplication(app)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
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
