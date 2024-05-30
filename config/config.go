package config

import "os"

var Version = "1.1.8"
var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")
var Prefix = os.Getenv("PREFIX")

var PostgresHost = os.Getenv("POSTGRES_HOST")
var PostgresUser = os.Getenv("POSTGRES_USER")
var PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
var PostgresPort = os.Getenv("POSTGRES_PORT")

var DiscordToken = os.Getenv("DISCORD_TOKEN")
var DiscordGuild = os.Getenv("DISCORD_GUILD")
var DiscordLogChannel = os.Getenv("DISCORD_LOG_CHANNEL")

var GithubToken = os.Getenv("GITHUB_PAT")

var AdminRoleID = "1030681203864522823"
var OfficerRoleID = "812948550819905546"
var LeadRoleID = "970423652791246888"
