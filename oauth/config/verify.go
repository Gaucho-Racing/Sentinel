package config

import (
	"os"
	"strconv"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
)

func Verify() {
	if Env == "" {
		Env = "PROD"
		logger.SugarLogger.Infof("ENV is not set, defaulting to %s", Env)
	}
	if Port == "" {
		Port = "9997"
		logger.SugarLogger.Infof("PORT is not set, defaulting to %s", Port)
	}
	if DatabaseHost == "" {
		DatabaseHost = "localhost"
		logger.SugarLogger.Infof("DATABASE_HOST is not set, defaulting to %s", DatabaseHost)
	}
	if DatabasePort == "" {
		DatabasePort = "5432"
		logger.SugarLogger.Infof("DATABASE_PORT is not set, defaulting to %s", DatabasePort)
	}
	if DatabaseUser == "" {
		DatabaseUser = "postgres"
		logger.SugarLogger.Infof("DATABASE_USER is not set, defaulting to %s", DatabaseUser)
	}
	if DatabasePassword == "" {
		DatabasePassword = "password"
		logger.SugarLogger.Infof("DATABASE_PASSWORD is not set, defaulting to %s", DatabasePassword)
	}
	if DatabaseName == "" {
		DatabaseName = "sentinel"
		logger.SugarLogger.Infof("DATABASE_NAME is not set, defaulting to %s", DatabaseName)
	}
	AccessTokenTTL = parseIntEnv("ACCESS_TOKEN_TTL", 30*60)
	RefreshTokenTTL = parseIntEnv("REFRESH_TOKEN_TTL", 7*24*60*60)
}

func parseIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		logger.SugarLogger.Infof("%s is not set, defaulting to %d", key, fallback)
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		logger.SugarLogger.Warnf("%s is not a valid integer (%q), defaulting to %d", key, raw, fallback)
		return fallback
	}
	return n
}
