package main

import (
	"github.com/gaucho-racing/sentinel/discord/api"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/rincon"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	rincon.Init(&config.Service, &config.Routes)
	database.Init()
	service.ConnectDiscord()
	service.InitializeBot()

	api.Run()
}
