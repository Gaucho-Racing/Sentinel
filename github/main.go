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
	logger.Init()
	cfg, err := config.Load()
	if err != nil {
		logger.SugarLogger.Fatal(err)
	}
	kerbecs.Init(cfg.KerbecsEndpoint, cfg.KerbecsUser, cfg.KerbecsPassword)
	client, err := service.NewGitHubClient(cfg)
	if err != nil {
		logger.SugarLogger.Fatal(err)
	}
	if err := sentinel.Bootstrap(config.InternalServiceName, cfg.BootstrapSecret); err != nil {
		logger.SugarLogger.Fatalf("bootstrap core service account: %v", err)
	}
	server := service.NewServer(cfg, client)
	go func() {
		server.RunReconcile(context.Background())
		ticker := time.NewTicker(cfg.SyncInterval)
		defer ticker.Stop()
		for range ticker.C {
			server.RunReconcile(context.Background())
		}
	}()
	logger.SugarLogger.Infof("GitHub service listening on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, api.Router(server)); err != nil {
		logger.SugarLogger.Fatal(err)
	}
}
