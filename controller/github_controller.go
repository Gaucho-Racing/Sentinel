package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	cron "github.com/robfig/cron/v3"
)

func GetGithubStatusForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "read:github"))
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
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "write:github"))
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

func RegisteGithubCronJob() {
	c := cron.New()
	entryID, err := c.AddFunc(config.GithubCron, func() {
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":alarm_clock: Starting github CRON Job")
		utils.SugarLogger.Infoln("Starting github CRON Job...")
		service.CleanGithubMembers()
		utils.SugarLogger.Infoln("Finished github CRON Job!")
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Finished github job!")
	})
	if err != nil {
		utils.SugarLogger.Errorln("Error registering CRON Job: " + err.Error())
		return
	}
	c.Start()
	utils.SugarLogger.Infoln("Registered CRON Job: " + strconv.Itoa(int(entryID)) + " scheduled with cron expression: " + config.GithubCron)
}
