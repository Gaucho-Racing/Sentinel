package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"

	"github.com/gin-gonic/gin"
)

func GetDriveStatusForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "read:drive"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	perm, err := service.GetDriveMemberPermission(config.SharedDriveID, user.Email)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	if perm != nil {
		c.JSON(http.StatusOK, perm)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "No permissions found for user with id: " + userID})
}

func AddUserToDrive(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "write:drive"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	role := "writer"
	if user.IsInnerCircle() {
		role = "organizer"
	}
	err := service.AddMemberToDrive(config.SharedDriveID, user.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User added to drive"})
}

func RemoveUserFromDrive(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "write:drive"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	userID := c.Param("userID")
	user := service.GetUserByID(userID)
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with id: " + userID})
		return
	}
	err := service.RemoveMemberFromDrive(config.SharedDriveID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User removed from drive"})
}
