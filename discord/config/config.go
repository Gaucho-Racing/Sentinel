package config

import (
	"os"
	"time"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "Sentinel Discord",
	Version:     "5.4.0",
	Endpoint:    os.Getenv("SERVICE_ENDPOINT"),
	HealthCheck: os.Getenv("SERVICE_HEALTH_CHECK"),
}

var Routes = []rincon.Route{
	{
		Route:  "/discord/**",
		Method: "*",
	},
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

var DiscordToken = os.Getenv("DISCORD_TOKEN")
var DiscordGuild = os.Getenv("DISCORD_GUILD")
var DiscordPrefix = os.Getenv("DISCORD_PREFIX")

var WebBaseURL = os.Getenv("WEB_BASE_URL")

var OnboardingTokenTTL = 15 * time.Minute

func IsProduction() bool {
	return Env == "PROD"
}

var MembersDiscordRoleID = "820467859477889034"
var AlumniDiscordRoleID = "817577502968512552"
var GuestDiscordRoleID = "1511273081824477245"

var AeroSubteamDiscordRoleID = "761114473565519882"
var BusinessSubteamDiscordRoleID = "761331962563919874"
var ChassisSubteamDiscordRoleID = "761114557531553824"
var DataSubteamDiscordRoleID = "1254572624307290202"
var DrivetrainSubteamDiscordRoleID = "1344560076765007893"
var ElectronicsSubteamDiscordRoleID = "761116347865890816"
var FirmwareSubteamDiscordRoleID = "1387486553483509921"
var SuspensionSubteamDiscordRoleID = "761114667048763423"