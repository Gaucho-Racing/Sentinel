package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"

	"github.com/gin-gonic/gin"
)

func RegisterAccountPassword(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

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
	RequireAny(c, RequestUserHasID(c, user.ID), RequestUserHasRole(c, "d_admin"))

	token, err := service.RegisterEmailPassword(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = service.CreateWikiUserWithPassword(input.Password, user.ID)
	if err != nil {
		utils.SugarLogger.Errorf("Error creating wiki user: %v", err)
		utils.SugarLogger.Infoln("Attempting to update wiki user")
		err = service.UpdateWikiUserWithPassword(input.Password, user.ID)
		if err != nil {
			utils.SugarLogger.Errorf("Error updating wiki user: %v", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "token": token})
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
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "token": token})
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
	token, err := service.GenerateJWT(user.ID, user.Email, "sentinel:all", "sentinel")
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
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "token": token})
}

func GetAuthForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

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
