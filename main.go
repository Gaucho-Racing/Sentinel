package main

import (
	"github.com/gin-gonic/gin"
	"sentinel/commands"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
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
	service.ConnectDiscord()
	commands.InitializeDiscordBot()

	router.Run(":" + config.Port)
}
