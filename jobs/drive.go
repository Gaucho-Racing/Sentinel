package jobs

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"
	"sync"

	cron "github.com/robfig/cron/v3"
)

func RegisterDriveCronJob() {
	if config.Env != "PROD" {
		utils.SugarLogger.Infoln("Drive CRON Job not registered because environment is not PROD")
		return
	}
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
