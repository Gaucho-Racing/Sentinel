package utils

import (
	"sentinel/config"
	"testing"
)

func TestVerifyConfig(t *testing.T) {
	InitializeLogger()
	t.Run("Test Blank Config", func(t *testing.T) {
		config.Env = ""
		config.Port = ""
		config.PostgresHost = ""
		config.PostgresPort = ""
		config.PostgresUser = ""
		config.PostgresPassword = ""
		config.DiscordToken = ""
		config.DiscordGuild = ""
		config.DiscordLogChannel = ""
		config.GithubToken = ""
		VerifyConfig()
	})
}
