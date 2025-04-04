package config

import (
	"crypto/rsa"
	"os"
)

var Version = "4.2.4"
var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")
var Prefix = os.Getenv("PREFIX")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

var DiscordToken = os.Getenv("DISCORD_TOKEN")
var DiscordGuild = os.Getenv("DISCORD_GUILD")
var DiscordLogChannel = os.Getenv("DISCORD_LOG_CHANNEL")

var DiscordClientID = os.Getenv("DISCORD_CLIENT_ID")
var DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
var DiscordRedirectURI = os.Getenv("DISCORD_REDIRECT_URI")

var DriveServiceAccount = os.Getenv("DRIVE_SERVICE_ACCOUNT")
var GithubToken = os.Getenv("GITHUB_PAT")
var WikiToken = os.Getenv("WIKI_TOKEN")

var SharedDriveID = "0ADMP93ZBlor_Uk9PVA"
var LeadsDriveID = "0AF4DbFL3cclkUk9PVA"

var AdminRoleID = "1030681203864522823"
var OfficerRoleID = "812948550819905546"
var LeadRoleID = "970423652791246888"
var MemberRoleID = "1334383074410237984"
var AlumniRoleID = "817577502968512552"
var BotRoleID = "1229611357259694132"

var SubteamRoleNames = []string{"Aero", "Business", "Chassis", "Data", "Electronics", "Powertrain", "Suspension"}

var RsaPublicKey *rsa.PublicKey
var RsaPrivateKey *rsa.PrivateKey
var RsaPublicKeyJWKS map[string]interface{}

var RsaPublicKeyString = os.Getenv("RSA_PUBLIC_KEY")
var RsaPrivateKeyString = os.Getenv("RSA_PRIVATE_KEY")

var MemberDirectorySheetID = "1reuLZox2daj8r2H-lZrwB4oFPYlJ6oC7983UUaZd6AY"

var DriveCron = os.Getenv("DRIVE_CRON")
var GithubCron = os.Getenv("GITHUB_CRON")
var DiscordCron = os.Getenv("DISCORD_CRON")
