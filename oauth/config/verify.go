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
	if Issuer == "" {
		Issuer = "https://sso.gauchoracing.com"
		logger.SugarLogger.Infof("ISSUER is not set, defaulting to %s", Issuer)
	}
	if KerbecsEndpoint == "" {
		KerbecsEndpoint = "http://localhost:10300"
		logger.SugarLogger.Infof("KERBECS_ENDPOINT is not set, defaulting to %s", KerbecsEndpoint)
	}
	if KerbecsUser == "" {
		KerbecsUser = "admin"
		logger.SugarLogger.Infof("KERBECS_USER is not set, defaulting to %s", KerbecsUser)
	}
	if KerbecsPassword == "" {
		KerbecsPassword = "admin"
		logger.SugarLogger.Infoln("KERBECS_PASSWORD is not set, defaulting to \"admin\" — DO NOT USE IN PRODUCTION")
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
