package main

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
)

func main() {
	config.PrintStartupBanner()
	utils.InitializeLogger()
	utils.VerifyConfig()
	defer utils.Logger.Sync()

	service.InitializeKeys()

	// database.InitializeDB()
	// service.InitializeDrive()
	// service.ConnectDiscord()
	// service.InitializeRoles()
	// service.InitializeSubteams()
	// go service.SyncRolesForAllUsers()
	// commands.InitializeDiscordBot()
	// controller.RegisterDriveCronJob()
	// controller.RegisteGithubCronJob()
	// controller.RegisterWikiCronJob()

	token, err := service.GenerateJWT("123", "test@test.com", "sentinel:all", "sentinel")
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}
	utils.SugarLogger.Infoln(token)

	claims, err := service.ValidateJWT(token)
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}
	utils.SugarLogger.Infoln(claims)

	// router := controller.SetupRouter()
	// controller.InitializeRoutes(router)
	// err := router.Run(":" + config.Port)
	// if err != nil {
	// 	utils.SugarLogger.Fatalln(err)
	// }
}
