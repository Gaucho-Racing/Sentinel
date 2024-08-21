package main

import (
	"sentinel/commands"
	"sentinel/config"
	"sentinel/controller"
	"sentinel/database"
	"sentinel/service"
	"sentinel/utils"
)

func main() {
	config.PrintStartupBanner()
	utils.InitializeLogger()
	utils.VerifyConfig()
	defer utils.Logger.Sync()

	database.InitializeDB()
	service.InitializeDrive()
	service.ConnectDiscord()
	service.InitializeRoles()
	service.InitializeSubteams()
	go service.SyncRolesForAllUsers()
	commands.InitializeDiscordBot()
	controller.RegisterDriveCronJob()
	controller.RegisteGithubCronJob()
	controller.RegisterWikiCronJob()
	// service.FindAllNonVerifiedUsers()

	c, e := service.GenerateAuthorizationCode("test", "test", "test")
	if e != nil {
		utils.SugarLogger.Errorln(e)
	}
	utils.SugarLogger.Info(c)

	router := controller.SetupRouter()
	controller.InitializeRoutes(router)
	err := router.Run(":" + config.Port)
	if err != nil {
		utils.SugarLogger.Fatalln(err)
	}
}
