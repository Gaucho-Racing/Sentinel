package main

import (
	"github.com/gaucho-racing/sentinel/core/api"
	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/pkg/rincon"
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
