package main

import (
	"sentinel/config"
	"sentinel/database"
	"sentinel/service"
	"sentinel/utils"
)

func main() {
	config.PrintStartupBanner()
	utils.InitializeLogger()
	defer utils.Logger.Sync()

	database.InitializeDB()
	service.InitializeDrive()
	service.TestDrive()
	// service.ConnectDiscord()
	// service.InitializeRoles()
	// service.InitializeSubteams()
	// go service.SyncRolesForAllUsers()
	// commands.InitializeDiscordBot()
	// // service.FindAllNonVerifiedUsers()

	// router := controller.SetupRouter()
	// controller.InitializeRoutes(router)
	// err := router.Run(":" + config.Port)
	// if err != nil {
	// 	utils.SugarLogger.Fatalln(err)
	// }
}
