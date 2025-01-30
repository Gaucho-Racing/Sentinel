package main

import (
	"sentinel/commands"
	"sentinel/config"
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
	service.InitializeKeys()
	service.InitializeDrive()
	service.ConnectDiscord()
	service.InitializeRoles()
	service.InitializeSubteams()
	go service.SyncRolesForAllUsers()
	commands.InitializeDiscordBot()

	service.PopulateMemberDirectorySheet()

	// jobs.RegisterDriveCronJob()
	// jobs.RegisteGithubCronJob()
	// jobs.RegisterDiscordCronJob()

	// router := controller.SetupRouter()
	// controller.InitializeRoutes(router)
	// err := router.Run(":" + config.Port)
	// if err != nil {
	// 	utils.SugarLogger.Fatalln(err)
	// }
}
