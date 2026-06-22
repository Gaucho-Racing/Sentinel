package config

import (
	"os"
)

const Name = "sentinel-oauth"
const Version = "5.6.3"

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

// TeamGoogleClientID is the client_id of the shared Google Workspace account
// application (team@gauchoracing.com). Identity-override lookups key on it, so
// it's read from the environment to stay correct across deployments. When
// unset, no override is registered for it.
var TeamGoogleClientID = os.Getenv("TEAM_GOOGLE_CLIENT_ID")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

// InternalBootstrapSecret is the shared secret this service uses at
// startup to exchange for its pre-seeded bearer JWT from core. Must
// match core's INTERNAL_BOOTSTRAP_SECRET.
var InternalBootstrapSecret = os.Getenv("INTERNAL_BOOTSTRAP_SECRET")

// InternalServiceName is the SA name on core that this service exchanges
// the bootstrap secret for. Must match a value in
// core/jobs/init.go::InternalServiceAccountNames.
const InternalServiceName = "sentinel-oauth"

// Discord OAuth for "Continue with Discord" on the login page. The redirect
// URI must byte-match the one the web client used in its authorize step —
// Discord rejects the token exchange otherwise — so the web reads its own
// VITE_ copy of this value and both must point at the same web callback
// route (typically <origin>/auth/login/discord).
var DiscordClientID     = os.Getenv("DISCORD_CLIENT_ID")
var DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
var DiscordRedirectURI  = os.Getenv("DISCORD_REDIRECT_URI")

func IsProduction() bool {
	return Env == "PROD"
}
