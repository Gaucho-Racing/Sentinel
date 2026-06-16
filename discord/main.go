package main

import (
	"github.com/gaucho-racing/sentinel/discord/api"
	"github.com/gaucho-racing/sentinel/discord/commands"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	kerbecs.Init(config.KerbecsEndpoint, config.KerbecsUser, config.KerbecsPassword)
	database.Init()
	service.ConnectDiscord()
	commands.InitializeBot()
	service.StartReconcileCron()

	api.Run()
}
