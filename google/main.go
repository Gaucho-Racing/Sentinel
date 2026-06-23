package main

import (
	"github.com/gaucho-racing/sentinel/google/api"
	"github.com/gaucho-racing/sentinel/google/config"
	"github.com/gaucho-racing/sentinel/google/database"
	"github.com/gaucho-racing/sentinel/google/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
	"github.com/gaucho-racing/sentinel/google/pkg/sentinel"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	kerbecs.Init(config.KerbecsEndpoint, config.KerbecsUser, config.KerbecsPassword)

	// Exchange the shared bootstrap secret for this service's pre-seeded
	// bearer JWT. From here on, every outbound sentinel-client call
	// carries Authorization: Bearer <our SA token>.
	if err := sentinel.Bootstrap(config.InternalServiceName, config.InternalBootstrapSecret); err != nil {
		logger.SugarLogger.Fatalf("Failed to bootstrap service token: %v", err)
	}

	database.Init()

	api.Run()
}
