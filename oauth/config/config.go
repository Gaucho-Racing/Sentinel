package config

import (
	"os"
)

const Name = "sentinel-oauth"
const Version = "5.4.3"

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
