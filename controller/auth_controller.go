package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func RegisterAccount(c *gin.Context) {
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
	token, err := service.RegisterEmailPassword(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
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
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "token": token})
}

func LoginDiscord(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No code provided"})
		return
	}
	println(code)
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
	token, err := service.GenerateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "token": token})
}

func GetAuthForUser(c *gin.Context) {
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
