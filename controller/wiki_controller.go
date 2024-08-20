package controller

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"

	cron "github.com/robfig/cron/v3"
)

func RegisterWikiCronJob() {
	c := cron.New()
	entryID, err := c.AddFunc(config.WikiCron, func() {
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":alarm_clock: Starting wiki CRON Job")
		utils.SugarLogger.Infoln("Starting wiki CRON Job...")
		service.CleanWikiMembers()
		utils.SugarLogger.Infoln("Finished wiki CRON Job!")
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Finished wiki job!")
	})
	if err != nil {
		utils.SugarLogger.Errorln("Error registering CRON Job: " + err.Error())
		return
	}
	c.Start()
	utils.SugarLogger.Infoln("Registered CRON Job: " + strconv.Itoa(int(entryID)) + " scheduled with cron expression: " + config.WikiCron)
}
