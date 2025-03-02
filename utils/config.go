package utils

import (
	"sentinel/config"
	"strings"
)

func VerifyConfig() {
	if config.Port == "" {
		config.Port = "9999"
		SugarLogger.Infof("PORT is not set, defaulting to %s", config.Port)
	}
	if config.DatabaseHost == "" {
		config.DatabaseHost = "localhost"
		SugarLogger.Infof("DATABASE_HOST is not set, defaulting to %s", config.DatabaseHost)
	}
	if config.DatabasePort == "" {
		config.DatabasePort = "3306"
		SugarLogger.Infof("DATABASE_PORT is not set, defaulting to %s", config.DatabasePort)
	}
	if config.DatabaseUser == "" {
		config.DatabaseUser = "root"
		SugarLogger.Infof("DATABASE_USER is not set, defaulting to %s", config.DatabaseUser)
	}
	if config.DatabasePassword == "" {
		config.DatabasePassword = "password"
		SugarLogger.Infof("DATABASE_PASSWORD is not set, defaulting to %s", config.DatabasePassword)
	}
	if config.DiscordToken == "" {
		SugarLogger.Fatalf("DISCORD_TOKEN is not set")
	}
	if config.DiscordGuild == "" {
		SugarLogger.Fatalf("DISCORD_GUILD is not set")
	}
	if config.DiscordLogChannel == "" {
		SugarLogger.Fatalf("DISCORD_LOG_CHANNEL is not set")
	}
	if config.GithubToken == "" {
		SugarLogger.Fatalf("GITHUB_PAT is not set")
	}
	if config.DriveServiceAccount == "" {
		SugarLogger.Fatalf("DRIVE_SERVICE_ACCOUNT is not set")
	}
	if config.WikiToken == "" {
		SugarLogger.Fatalf("WIKI_TOKEN is not set")
	}
	if config.RsaPublicKeyString == "" {
		SugarLogger.Fatalf("RSA_PUBLIC_KEY is not set")
	}
	config.RsaPublicKeyString = strings.ReplaceAll(config.RsaPublicKeyString, "\\n", "\n")
	if config.RsaPrivateKeyString == "" {
		SugarLogger.Fatalf("RSA_PRIVATE_KEY is not set")
	}
	config.RsaPrivateKeyString = strings.ReplaceAll(config.RsaPrivateKeyString, "\\n", "\n")
	if config.DriveCron == "" {
		config.DriveCron = "0 * * * *"
		SugarLogger.Infof("DRIVE_CRON is not set, defaulting to %s", config.DriveCron)
	}
	if config.GithubCron == "" {
		config.GithubCron = "0 * * * *"
		SugarLogger.Infof("GITHUB_CRON is not set, defaulting to %s", config.GithubCron)
	}
	if config.DiscordCron == "" {
		config.DiscordCron = "0 * * * *"
		SugarLogger.Infof("DISCORD_CRON is not set, defaulting to %s", config.DiscordCron)
	}
}
