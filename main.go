package main

import (
	"sentinel/commands"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func setupRouter() *gin.Engine {
	if config.Env == "PROD" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	return r
}

func main() {
	utils.InitializeLogger()
	defer utils.Logger.Sync()

	router = setupRouter()
	service.InitializeDB()
	service.InitializeSubteams()
	service.ConnectDiscord()
	commands.InitializeDiscordBot()

	service.FindAllNonVerifiedUsers()

	router.Run(":" + config.Port)
}
