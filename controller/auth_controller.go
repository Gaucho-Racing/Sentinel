package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"

	"github.com/gin-gonic/gin"
)

func GetJWKS(c *gin.Context) {
	c.JSON(http.StatusOK, config.RsaPublicKeyJWKS)
}

func RegisterAccountPassword(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	var input model.UserAuth
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := service.GetUserByEmail(input.Email)
	if user.ID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "No account with this email exists. Make sure to verify your account on the discord server first!"})
		return
	}

	Require(c, Any(RequestUserHasID(c, user.ID), RequestUserHasRole(c, "d_admin")))

	token, err := service.RegisterEmailPassword(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	refreshToken := ""
	response := model.TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    24 * 60,
		Scope:        "sentinel:all",
	}
	c.JSON(http.StatusOK, response)
}

func ResetAccountPassword(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}

	Require(c, Any(RequestUserHasID(c, user.ID), RequestUserHasRole(c, "d_admin")))

	auth := service.GetUserAuthByID(userID)
	if auth.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No authentication found for user with id: " + userID})
		return
	}

	err := service.RemovePasswordForEmail(auth.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func LoginAccount(c *gin.Context) {
	var input model.UserAuth
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := service.GetUserByEmail(input.Email)
	if user.ID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "No account with this email exists. Make sure to verify your account on the discord server first!"})
		return
	}
	token, err := service.LoginEmailPassword(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	go service.CreateLogin(model.UserLogin{
		UserID:      user.ID,
		Destination: "sentinel",
		Scope:       "sentinel:all",
		IPAddress:   c.ClientIP(),
		LoginType:   "email",
	})
	refreshToken, err := service.GenerateRefreshToken(user.ID, "sentinel:all", "sentinel", 7*24*60*60)
	if err != nil {
		utils.SugarLogger.Errorln("error generating refresh token: " + err.Error())
		refreshToken = ""
	}
	response := model.TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    24 * 60,
		Scope:        "sentinel:all",
	}
	c.JSON(http.StatusOK, response)
}

func LoginDiscord(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No code provided"})
		return
	}
	id, err := service.GetUserIDFromDiscordCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	user := service.GetUserByID(id)
	if user.ID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "No account with this email exists. Make sure to verify your account on the discord server first!"})
		return
	}
	token, err := service.GenerateAccessToken(user.ID, "sentinel:all", "sentinel", 24*60*60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	go service.CreateLogin(model.UserLogin{
		UserID:      user.ID,
		Destination: "sentinel",
		Scope:       "sentinel:all",
		IPAddress:   c.ClientIP(),
		LoginType:   "discord",
	})
	refreshToken, err := service.GenerateRefreshToken(user.ID, "sentinel:all", "sentinel", 7*24*60*60)
	if err != nil {
		utils.SugarLogger.Errorln("error generating refresh token: " + err.Error())
		refreshToken = ""
	}
	response := model.TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    24 * 60,
		Scope:        "sentinel:all",
	}
	c.JSON(http.StatusOK, response)
}

func GetAuthForUser(c *gin.Context) {
	Require(c, All(
		RequestTokenHasScope(c, "sentinel:all"),
		Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
	))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	auth := service.GetUserAuthByID(userID)
	if auth.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No authentication found for user with id: " + userID})
		return
	}
	auth.Password = "************"
	c.JSON(http.StatusOK, auth)
}
