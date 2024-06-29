package config

import "os"

var Version = "2.0.0"
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

var DiscordClientID = os.Getenv("DISCORD_CLIENT_ID")
var DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
var DiscordRedirectURI = os.Getenv("DISCORD_REDIRECT_URI")

var DriveServiceAccount = os.Getenv("DRIVE_SERVICE_ACCOUNT")
var GithubToken = os.Getenv("GITHUB_PAT")

var SharedDriveID = "0ADMP93ZBlor_Uk9PVA"
var LeadsDriveID = "0AF4DbFL3cclkUk9PVA"

var AdminRoleID = "1030681203864522823"
var OfficerRoleID = "812948550819905546"
var LeadRoleID = "970423652791246888"
var MemberRoleID = "820467859477889034"
var AlumniRoleID = "817577502968512552"

var SubteamRoleNames = []string{"Aero", "Business", "Chassis", "Data", "Electronics", "Powertrain", "Suspension"}

var AuthSigningKey = os.Getenv("AUTH_SIGNING_KEY")
