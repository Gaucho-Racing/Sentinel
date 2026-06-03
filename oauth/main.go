package main

import (
	"github.com/gaucho-racing/sentinel/oauth/api"
	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/database"
	"github.com/gaucho-racing/sentinel/oauth/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	kerbecs.Init(config.KerbecsEndpoint, config.KerbecsUser, config.KerbecsPassword)
	database.Init()

	api.Run()
}
