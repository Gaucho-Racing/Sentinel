package main

import (
	"github.com/gaucho-racing/sentinel/discord/api"
	"github.com/gaucho-racing/sentinel/discord/commands"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/discord/service"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	kerbecs.Init(config.KerbecsEndpoint, config.KerbecsUser, config.KerbecsPassword)

	// Exchange the shared bootstrap secret for this service's pre-seeded
	// bearer JWT. From here on, every outbound sentinel-client call
	// carries Authorization: Bearer <our SA token>. Fatal on failure —
	// compose's restart: always handles the boot race after a few
	// minutes if it ever persists.
	if err := sentinel.Bootstrap(config.InternalServiceName, config.InternalBootstrapSecret); err != nil {
		logger.SugarLogger.Fatalf("Failed to bootstrap service token: %v", err)
	}

	database.Init()
	service.ConnectDiscord()
	commands.InitializeBot()
	service.StartReconcileCron()

	api.Run()
}
