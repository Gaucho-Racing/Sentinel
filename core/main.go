package main

import (
	"github.com/gaucho-racing/sentinel/core/api"
	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/jobs"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	database.Init()
	service.InitializeKeys()
	jobs.InitializeCore()

	// Initial conditional-group reconcile to catch drift accumulated while
	// core was offline (manual DB edits, missed triggers, etc). Background
	// goroutine via syncJob so it doesn't block startup.
	service.TriggerReconcileAllConditional()
	// Periodic safety-net sweep on a configurable interval.
	service.StartReconcileConditionalCron()

	api.Run()
}
