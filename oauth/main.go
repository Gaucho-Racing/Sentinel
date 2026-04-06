package main

import (
	"github.com/gaucho-racing/sentinel/oauth/api"
	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/database"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/rincon"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	rincon.Init(&config.Service, &config.Routes)
	database.Init()

	api.Run()
}
