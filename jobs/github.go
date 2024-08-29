package jobs

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"

	cron "github.com/robfig/cron/v3"
)

func RegisteGithubCronJob() {
	if config.Env != "PROD" {
		utils.SugarLogger.Infoln("Github CRON Job not registered because environment is not PROD")
		return
	}
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
