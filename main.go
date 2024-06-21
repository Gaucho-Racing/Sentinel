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
	utils.InitializeLogger()
	defer utils.Logger.Sync()

	database.InitializeDB()
	service.InitializeSubteams()
	service.ConnectDiscord()
	commands.InitializeDiscordBot()
	// service.FindAllNonVerifiedUsers()

	router := controller.SetupRouter()
	controller.InitializeRoutes(router)
	err := router.Run(":" + config.Port)
	if err != nil {
		utils.SugarLogger.Fatalln(err)
	}
}
