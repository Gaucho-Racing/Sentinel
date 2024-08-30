package jobs

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"

	cron "github.com/robfig/cron/v3"
)

func RegisteDiscordCronJob() {
	if config.Env != "PROD" {
		utils.SugarLogger.Infoln("Discord CRON Job not registered because environment is not PROD")
		return
	}
	c := cron.New()
	CleanDiscordJob(c)
	IncompleteProfileJob(c)
}

func CleanDiscordJob(c *cron.Cron) {
	entryID, err := c.AddFunc(config.DiscordCron, func() {
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":alarm_clock: Starting discord CRON Job")
		utils.SugarLogger.Infoln("Starting discord CRON Job...")
		service.CleanGithubMembers()
		utils.SugarLogger.Infoln("Finished discord CRON Job!")
		_, _ = service.Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Finished discord job!")
	})
	if err != nil {
		utils.SugarLogger.Errorln("Error registering CRON Job: " + err.Error())
		return
	}
	c.Start()
	utils.SugarLogger.Infoln("Registered CRON Job: " + strconv.Itoa(int(entryID)) + " scheduled with cron expression: " + config.DiscordCron)
}

func IncompleteProfileJob(c *cron.Cron) {

}
