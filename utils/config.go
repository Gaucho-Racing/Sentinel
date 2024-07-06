package utils

import "sentinel/config"

func VerifyConfig() {
	if config.Port == "" {
		config.Port = "7999"
		SugarLogger.Infof("PORT is not set, defaulting to %s", config.Port)
	}
	if config.PostgresHost == "" {
		config.PostgresHost = "localhost"
		SugarLogger.Infof("POSTGRES_HOST is not set, defaulting to %s", config.PostgresHost)
	}
	if config.PostgresPort == "" {
		config.PostgresPort = "5432"
		SugarLogger.Infof("POSTGRES_PORT is not set, defaulting to %s", config.PostgresPort)
	}
	if config.PostgresUser == "" {
		config.PostgresUser = "postgres"
		SugarLogger.Infof("POSTGRES_USER is not set, defaulting to %s", config.PostgresUser)
	}
	if config.PostgresPassword == "" {
		config.PostgresPassword = "postgres"
		SugarLogger.Infof("POSTGRES_PASSWORD is not set, defaulting to %s", config.PostgresPassword)
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
}
