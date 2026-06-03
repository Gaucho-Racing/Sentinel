package config

import (
	"os"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "Sentinel OAuth",
	Version:     "5.4.3",
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

// Issuer is the OIDC issuer identifier and the public base URL of the
// deployment. It MUST match core's ISSUER exactly — the discovery document's
// `issuer` and the `iss` claim core stamps into ID tokens have to be
// byte-identical for relying parties to accept the token.
var Issuer = os.Getenv("ISSUER")

var AccessTokenTTL int
var RefreshTokenTTL int

// SentinelClientID is the first-party identifier used by /auth/login and
// /auth/refresh when minting tokens. Direct-login tokens carry this as
// their client_id, which is how token issuance distinguishes Sentinel's
// own session tokens from third-party OAuth tokens. Group links attached
// to the matching application row act as a global default that's added
// to every OAuth app's filter set.
const SentinelClientID = "sentinel"

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

func IsProduction() bool {
	return Env == "PROD"
}
