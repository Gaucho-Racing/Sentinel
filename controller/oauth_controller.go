package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func GetValidOauthScopes(c *gin.Context) {
	c.JSON(http.StatusOK, model.ValidOauthScopes)
}

func GetOpenIDConfig(c *gin.Context) {
	c.JSON(http.StatusOK, model.OpenIDConfig)
}

func GetAllClientApplications(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "applications:read"),
			RequestUserHasRole(c, "d_admin"),
		),
	))

	apps := service.GetAllClientApplications()
	c.JSON(http.StatusOK, apps)
}

func GetClientApplicationsForUser(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "applications:read"),
			Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
		),
	))

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

	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "applications:read"),
			Any(RequestUserHasID(c, app.UserID), RequestUserHasRole(c, "d_admin")),
		),
	))

	c.JSON(http.StatusOK, app)
}

func CreateClientApplication(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	var app model.ClientApplication
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if app.ID != "" {
		existing := service.GetClientApplicationByID(app.ID)
		Require(c, Any(
			RequestUserHasID(c, existing.UserID),
			RequestUserHasRole(c, "d_admin"),
		))
	} else {
		app.UserID = GetRequestUserID(c)
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
	app := service.GetClientApplicationByID(appID)
	if app.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "no client application found with id: " + appID})
		return
	}

	Require(c, All(
		RequestTokenHasScope(c, "sentinel:all"),
		Any(RequestUserHasID(c, app.UserID), RequestUserHasRole(c, "d_admin")),
	))

	err := service.DeleteClientApplication(appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client application deleted"})
}

func OauthAuthorize(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

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
	} else if !service.ValidateScope(scope) || strings.Contains(scope, "sentinel:all") {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scope is invalid"})
		return
	}
	// there seems to be a variety of prompts that people send as part of the oauth flow
	// the oidc spec says to prompt when prompt=login and bypass when prompt=none
	// discord prompts when prompt=consent and bypass when prompt=none
	// portainer prompts when prompt=login and just doesn't send the prompt at all when bypassing
	//
	// We will first check if there is no prompt provided, defaulting to requiring consent (prompt=consent)
	// If the prompt is set to none, we check if the user has previously authorized this client
	// if any other prompt is provided, we will default to requiring consent (prompt=consent)
	prompt := c.Query("prompt")
	if prompt == "" {
		prompt = "consent"
	}
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
	reponseType := c.Query("response_type")
	if reponseType == "" {
		reponseType = "code"
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
	defer service.CreateLogin(model.UserLogin{
		UserID:      GetRequestUserID(c),
		Destination: clientID,
		Scope:       scope,
		IPAddress:   c.ClientIP(),
		LoginType:   "oauth",
	})
	if reponseType == "code" {
		code, err := service.GenerateAuthorizationCode(clientID, GetRequestUserID(c), scope)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, code)
		return
	}
}

func OauthExchange(c *gin.Context) {
	// Check if refresh or authorization code
	grantType := c.PostForm("grant_type")
	if grantType == "refresh_token" {
		handleRefreshTokenExchange(c)
		return
	}
	// Check for Basic Auth
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if hasAuth {
		// Validate client credentials
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" || client.Secret != clientSecret {
			utils.SugarLogger.Errorf("invalid client credentials: %s %s", clientID, clientSecret)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid client credentials"})
			return
		}
	} else {
		// Check for client_id and client_secret in form
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
		if clientID == "" || clientSecret == "" {
			utils.SugarLogger.Errorf("client_id and client_secret are required: %s %s", clientID, clientSecret)
			c.JSON(http.StatusBadRequest, gin.H{"message": "client_id and client_secret are required"})
			return
		}
		// Validate client credentials
		client := service.GetClientApplicationByID(clientID)
		if client.ID == "" || client.Secret != clientSecret {
			utils.SugarLogger.Errorf("invalid client credentials: %s %s", clientID, clientSecret)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid client credentials"})
			return
		}
	}

	redirectUri := c.PostForm("redirect_uri")
	if !service.ValidateRedirectURI(redirectUri, clientID) {
		utils.SugarLogger.Errorf("redirect_uri is invalid: %s", redirectUri)
		c.JSON(http.StatusBadRequest, gin.H{"message": "redirect_uri is invalid"})
		return
	}
	if grantType == "" {
		utils.SugarLogger.Errorf("grant_type is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "grant_type is required"})
		return
	}
	if grantType == "authorization_code" {
		handleAuthorizationCodeExchange(c)
		return
	} else {
		utils.SugarLogger.Errorf("unsupported grant_type: %s", grantType)
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported grant_type"})
	}
}

func handleAuthorizationCodeExchange(c *gin.Context) {
	code := c.PostForm("code")
	if code == "" {
		utils.SugarLogger.Errorf("code is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "code is required"})
		return
	}
	authCode, err := service.VerifyAuthorizationCode(code)
	if err != nil {
		utils.SugarLogger.Errorf("error verifying authorization code: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	token, err := service.GenerateAccessToken(authCode.UserID, authCode.Scope, authCode.ClientID, 60*60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	idToken := ""
	if strings.Contains(authCode.Scope, "openid") {
		idToken, err = service.GenerateIDToken(authCode.UserID, authCode.Scope, authCode.ClientID, 60*60)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}
	refreshToken, err := service.GenerateRefreshToken(authCode.UserID, authCode.Scope, authCode.ClientID, 7*24*60*60)
	if err != nil {
		utils.SugarLogger.Errorln("error generating refresh token: " + err.Error())
		refreshToken = ""
	}
	response := model.TokenResponse{
		IDToken:      idToken,
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    60 * 60,
		Scope:        authCode.Scope,
	}
	utils.SugarLogger.Infof("token response: %v", response)
	c.JSON(http.StatusOK, response)
}

func handleRefreshTokenExchange(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		utils.SugarLogger.Errorf("refresh_token is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "refresh_token is required"})
		return
	}
	if !service.ValidateRefreshToken(refreshToken) {
		utils.SugarLogger.Errorf("invalid refresh_token: %s", refreshToken)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired refresh_token"})
		return
	}
	claims := &model.AuthClaims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return config.RsaPublicKey, nil
	})
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid refresh token"})
		return
	}
	if !strings.Contains(claims.Scope, "refresh_token") {
		utils.SugarLogger.Errorf("refresh token scope is required")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "provided token is not a refresh token"})
		return
	}
	go service.RevokeRefreshToken(refreshToken)
	// Remove refresh_token from scope
	scopeList := strings.Split(claims.Scope, " ")
	filteredScopes := make([]string, 0)
	for _, s := range scopeList {
		if s != "refresh_token" {
			filteredScopes = append(filteredScopes, s)
		}
	}
	claims.Scope = strings.Join(filteredScopes, " ")

	token, err := service.GenerateAccessToken(claims.Subject, claims.Scope, claims.Audience[0], 60*60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	idToken := ""
	if strings.Contains(claims.Scope, "openid") {
		idToken, err = service.GenerateIDToken(claims.Subject, claims.Scope, claims.Audience[0], 60*60)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}
	refreshToken, err = service.GenerateRefreshToken(claims.Subject, claims.Scope, claims.Audience[0], 7*24*60*60)
	if err != nil {
		utils.SugarLogger.Errorln("error generating refresh token: " + err.Error())
		refreshToken = ""
	}
	response := model.TokenResponse{
		IDToken:      idToken,
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    60 * 60,
		Scope:        claims.Scope,
	}
	utils.SugarLogger.Infof("token response: %v", response)
	c.JSON(http.StatusOK, response)
}
