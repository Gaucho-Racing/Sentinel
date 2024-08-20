package utils

import "sentinel/config"

func VerifyConfig() {
	if config.Port == "" {
		config.Port = "7999"
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
		SugarLogger.Errorf("DISCORD_TOKEN is not set")
	}
	if config.DiscordGuild == "" {
		SugarLogger.Errorf("DISCORD_GUILD is not set")
	}
	if config.DiscordLogChannel == "" {
		SugarLogger.Errorf("DISCORD_LOG_CHANNEL is not set")
	}
	if config.GithubToken == "" {
		SugarLogger.Errorf("GITHUB_PAT is not set")
	}
	if config.DriveServiceAccount == "" {
		SugarLogger.Errorf("DRIVE_SERVICE_ACCOUNT is not set")
	}
	if config.WikiToken == "" {
		SugarLogger.Errorf("WIKI_TOKEN is not set")
	}
	if config.AuthSigningKey == "" {
		SugarLogger.Errorf("AUTH_SIGNING_KEY is not set")
	}
	if config.DriveCron == "" {
		config.DriveCron = "0 * * * *"
		SugarLogger.Infof("DRIVE_CRON is not set, defaulting to %s", config.DriveCron)
	}
	if config.GithubCron == "" {
		config.GithubCron = "0 * * * *"
		SugarLogger.Infof("GITHUB_CRON is not set, defaulting to %s", config.GithubCron)
	}
}
