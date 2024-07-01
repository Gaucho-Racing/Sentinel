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
	// Check if clientID and clientSecret are in basic auth
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if hasAuth {
		// Verify the client credentials
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" || client.Secret != clientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid client credentials"})
			return
		}
	} else {
		// Check if clientID is in query params
		clientID = c.Query("client_id")
		clientSecret = c.Query("client_secret")
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" ||
	
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "client_id is required"})
		return
	}
	client := service.GetClientApplicationByID(clientID)
	if client.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "no client application found with id: " + clientID})
		return
	}
	c.HTML(http.StatusOK, "oauth_authorize.html", gin.H{
		"client": client,
	})
}
