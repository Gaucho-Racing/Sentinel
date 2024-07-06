package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	cron "github.com/robfig/cron/v3"
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

func RegisterDriveCronJob() {
	c := cron.New()
	entryID, err := c.AddFunc(config.DriveCron, func() {
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":alarm_clock: Starting google drive CRON Job")
		utils.SugarLogger.Infoln("Starting google drive CRON Job...")
		var wg sync.WaitGroup
		wg.Add(2)
		go service.PopulateMemberDirectorySheet()
		go service.CleanDriveMembers()
		wg.Wait()
		utils.SugarLogger.Infoln("Finished google drive CRON Job!")
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Finished google drive job!")
	})
	if err != nil {
		utils.SugarLogger.Errorln("Error registering CRON Job: " + err.Error())
		return
	}
	c.Start()
	utils.SugarLogger.Infoln("Registered CRON Job: " + strconv.Itoa(int(entryID)) + " scheduled with cron expression: " + config.DriveCron)
}
