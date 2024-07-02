package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func GetAllClientApplications(c *gin.Context) {
	apps := service.GetAllClientApplications()
	c.JSON(http.StatusOK, apps)
}

func GetClientApplicationsForUser(c *gin.Context) {
	userID := c.Param("userID")
	apps := service.GetClientApplicationsForUser(userID)
	c.JSON(http.StatusOK, apps)
}

func GetClientApplicationByID(c *gin.Context) {
	appID := c.Param("appID")
	app := service.GetClientApplicationByID(appID)
	if app.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "no client application found with id: " + appID})
		return
	}
	c.JSON(http.StatusOK, app)
}

func CreateClientApplication(c *gin.Context) {
	var app model.ClientApplication
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	created, err := service.CreateClientApplication(app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func DeleteClientApplication(c *gin.Context) {
	appID := c.Param("appID")
	err := service.DeleteClientApplication(appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client application deleted"})
}

func OauthAuthorize(c *gin.Context) {
	if service.GetRequestUserID(c) == "" || !service.RequestTokenHasScope(c, "sentinel:all") {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}

	clientID := c.Query("client_id")
	println(clientID)
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "client_id is required"})
		return
	}
	client := service.GetClientApplicationByID(clientID)
	if client.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no client application found with id: " + clientID})
		return
	}
	redirectUri := c.Query("redirect_uri")
	println(redirectUri)
	if !service.ValidateRedirectURI(redirectUri, clientID) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "redirect_uri is invalid"})
		return
	}
	scopes := c.Query("scopes")
	println(scopes)
	if scopes == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scopes are required"})
		return
	} else if !service.ValidateScopes(scopes) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scopes are invalid"})
		return
	}
	// Handle Validate Request
	if c.Request.Method == "GET" {
		c.JSON(http.StatusOK, gin.H{
			"client_id":    clientID,
			"redirect_uri": redirectUri,
			"scopes":       scopes,
		})
		return
	}
	// Handle Authorize Request
}
