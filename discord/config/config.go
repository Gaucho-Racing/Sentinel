package config

import (
	"os"
	"time"
)

const Name = "sentinel-discord"
const Version = "5.6.0"

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

// Kerbecs admin API — the gateway doubles as the service registry. The sentinel
// client resolves gateway-form paths (/api/core/...) to upstream URLs via its
// /admin-gw/resolve endpoint, which sits behind basic auth.
var KerbecsEndpoint = os.Getenv("KERBECS_ENDPOINT")
var KerbecsUser = os.Getenv("KERBECS_USER")
var KerbecsPassword = os.Getenv("KERBECS_PASSWORD")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

var DiscordToken = os.Getenv("DISCORD_TOKEN")
var DiscordGuild = os.Getenv("DISCORD_GUILD")
var DiscordPrefix = os.Getenv("DISCORD_PREFIX")

var WebBaseURL = os.Getenv("WEB_BASE_URL")

// InternalBootstrapSecret is the shared secret this service uses at
// startup to exchange for its pre-seeded bearer JWT from core. Must
// match core's INTERNAL_BOOTSTRAP_SECRET.
var InternalBootstrapSecret = os.Getenv("INTERNAL_BOOTSTRAP_SECRET")

// InternalServiceName is the SA name on core that this service exchanges
// the bootstrap secret for. Must match a value in
// core/jobs/init.go::InternalServiceAccountNames.
const InternalServiceName = "sentinel-discord"

var OnboardingTokenTTL = 15 * time.Minute

// GroupSyncInterval is how often the periodic reconcile cron fires. Event-
// driven sync covers the happy path; the cron is a safety net for missed
// gateway events (disconnects, restarts) and out-of-band changes that don't
// emit an event (e.g. group allowed_sources flips on the core side). Parsed
// once at startup via time.ParseDuration; an unparseable or unset value
// falls back to 1h. Set to 0 (or any non-positive duration) to disable.
var GroupSyncInterval = parseDurationOr("GROUP_SYNC_INTERVAL", time.Hour)

func parseDurationOr(envKey string, fallback time.Duration) time.Duration {
	raw := os.Getenv(envKey)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}

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
