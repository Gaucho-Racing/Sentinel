package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/github/api"
	"github.com/gaucho-racing/sentinel/github/config"
	"github.com/gaucho-racing/sentinel/github/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/github/pkg/logger"
	"github.com/gaucho-racing/sentinel/github/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/github/service"
)

func main() {
	logger.Init(config.Env == "PROD")
	defer logger.Logger.Sync()
	cfg, err := config.Load()
	if err != nil {
		logger.SugarLogger.Fatal(err)
	}
	config.PrintStartupBanner()
	kerbecs.Init(cfg.KerbecsEndpoint, cfg.KerbecsUser, cfg.KerbecsPassword)
	client, err := service.NewGitHubClient(cfg)
	if err != nil {
		logger.SugarLogger.Fatal(err)
	}
	if err := sentinel.Bootstrap(config.InternalServiceName, cfg.BootstrapSecret); err != nil {
		logger.SugarLogger.Fatalf("bootstrap core service account: %v", err)
	}
	server := service.NewServer(cfg, client)
	logger.SugarLogger.Infof("GitHub org sync: enabled, interval=%v", cfg.SyncInterval)
	go func() {
		logger.SugarLogger.Infof("GitHub org sync: kicking startup sweep")
		server.RunReconcile(context.Background())
		ticker := time.NewTicker(cfg.SyncInterval)
		defer ticker.Stop()
		for range ticker.C {
			logger.SugarLogger.Infof("GitHub org sync: cron tick, kicking full sweep")
			server.RunReconcile(context.Background())
		}
	}()
	logger.SugarLogger.Infof("GitHub service listening on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, api.Router(server)); err != nil {
		logger.SugarLogger.Fatal(err)
	}
}
