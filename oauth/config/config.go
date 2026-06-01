package config

import (
	"os"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "Sentinel OAuth",
	Version:     "5.0.0",
	Endpoint:    os.Getenv("SERVICE_ENDPOINT"),
	HealthCheck: os.Getenv("SERVICE_HEALTH_CHECK"),
}

var Routes = []rincon.Route{
	{
		Route:  "/oauth/**",
		Method: "*",
	},
	{
		Route:  "/auth/**",
		Method: "*",
	},
	{
		Route:  "/.well-known/**",
		Method: "GET",
	},
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

var AccessTokenTTL int
var RefreshTokenTTL int

// SentinelClientID identifies the first-party Sentinel client. Tokens minted
// for this client include the user's full group set (no filter); group links
// on this app act as a global default that's added to every other app's
// filter set during token issuance.
var SentinelClientID = os.Getenv("SENTINEL_CLIENT_ID")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

func IsProduction() bool {
	return Env == "PROD"
}
