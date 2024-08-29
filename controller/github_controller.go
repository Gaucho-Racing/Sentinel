package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func GetGithubStatusForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "github:read"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	github, err := service.GetGithubStatusForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, github)
}

func AddUserToGithub(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "github:write"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	var input model.GithubInvite
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if input.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Username is required"})
		return
	}
	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	err := service.AddUserToGithub(userID, input.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User added to github"})
}
