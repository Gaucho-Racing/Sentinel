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
		config.DatabaseHost = ""
		config.DatabasePort = ""
		config.DatabaseUser = ""
		config.DatabasePassword = ""
		config.DiscordToken = ""
		config.DiscordGuild = ""
		config.DiscordLogChannel = ""
		config.GithubToken = ""
		VerifyConfig()
	})
}
