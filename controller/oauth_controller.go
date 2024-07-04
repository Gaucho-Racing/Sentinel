package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllClientApplications(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	apps := service.GetAllClientApplications()
	c.JSON(http.StatusOK, apps)
}

func GetClientApplicationsForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "read:applications"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

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

	if !RequestTokenHasScope(c, "sentinel:all") {
		RequireAny(c, RequestTokenHasScope(c, "read:applications"))
		RequireAny(c, RequestUserHasRole(c, "d_admin"), RequestUserHasID(c, app.UserID))
	}

	c.JSON(http.StatusOK, app)
}

func CreateClientApplication(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	var app model.ClientApplication
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if app.ID != "" {
		existing := service.GetClientApplicationByID(app.ID)
		RequireAny(c, RequestUserHasID(c, existing.UserID), RequestUserHasRole(c, "d_admin"))
	}

	app.UserID = GetRequestUserID(c)

	created, err := service.CreateClientApplication(app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func DeleteClientApplication(c *gin.Context) {
	appID := c.Param("appID")
	app := service.GetClientApplicationByID(appID)
	if app.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "no client application found with id: " + appID})
		return
	}

	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))
	RequireAny(c, RequestUserHasRole(c, "d_admin"), RequestUserHasID(c, app.UserID))

	err := service.DeleteClientApplication(appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client application deleted"})
}

func OauthAuthorize(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	clientID := c.Query("client_id")
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
	if !service.ValidateRedirectURI(redirectUri, clientID) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "redirect_uri is invalid"})
		return
	}
	scope := c.Query("scope")
	if scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scope is required"})
		return
	} else if !service.ValidateScopes(scope) || strings.Contains(scope, "sentinel:all") {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scope is invalid"})
		return
	}
	prompt := c.Query("prompt")
	if prompt == "none" {
		// check if user previously authorized this client
		lastLogin := service.GetLastLoginForUserToDestinationWithScopes(GetRequestUserID(c), clientID, scope)
		if lastLogin.ID != "" && time.Since(lastLogin.CreatedAt).Hours() < 24*7 {
			utils.SugarLogger.Infof("User %s previously authorized client %s with scope %s", GetRequestUserID(c), clientID, scope)
			prompt = "none"
		} else {
			prompt = "consent"
		}
	} else {
		prompt = "consent"
	}
	// Handle Validate Request
	if c.Request.Method == "GET" {
		c.JSON(http.StatusOK, gin.H{
			"client_id":    clientID,
			"redirect_uri": redirectUri,
			"scope":        scope,
			"prompt":       prompt,
		})
		return
	}
	// Handle Authorize Request
	code, err := service.GenerateAuthorizationCode(clientID, GetRequestUserID(c), scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	go service.CreateLogin(model.UserLogin{
		UserID:      GetRequestUserID(c),
		Destination: clientID,
		Scope:       scope,
		IPAddress:   c.ClientIP(),
		LoginType:   "oauth",
	})
	c.JSON(http.StatusOK, code)
}

func OauthExchange(c *gin.Context) {
	// Check for Basic Auth
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if hasAuth {
		// Validate client credentials
		println(clientID, clientSecret)
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" || client.Secret != clientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid client credentials"})
			return
		}
	} else {
		// Check for client_id and client_secret in form
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
		if clientID == "" || clientSecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "client_id and client_secret are required"})
			return
		}
		// Validate client credentials
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" || client.Secret != clientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid client credentials"})
			return
		}
	}

	clientID = c.PostForm("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "client_id is required"})
		return
	}
	client := service.GetClientApplicationByID(clientID)
	if client.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no client application found with id: " + clientID})
		return
	}
	redirectUri := c.PostForm("redirect_uri")
	if !service.ValidateRedirectURI(redirectUri, clientID) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "redirect_uri is invalid"})
		return
	}
	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "code is required"})
		return
	}
	grantType := c.PostForm("grant_type")
	if grantType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "grant_type is required"})
		return
	}
	if grantType == "authorization_code" {
		handleAuthorizationCodeExchange(c)
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported grant_type"})
	}
}

func handleAuthorizationCodeExchange(c *gin.Context) {
	code := c.PostForm("code")
	authCode, err := service.VerifyAuthorizationCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	token, err := service.GenerateJWT(authCode.UserID, service.GetUserByID(authCode.UserID).Email, authCode.Scope, authCode.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	response := model.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   24 * 60 * 60,
		Scope:       authCode.Scope,
	}
	c.JSON(http.StatusOK, response)
}
