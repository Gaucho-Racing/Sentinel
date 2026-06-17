package config

import (
	"crypto/rsa"
	"os"
	"time"
)

const Name = "sentinel-core"
const Version = "5.6.0"

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

// Issuer is the OIDC issuer identifier. It MUST be a byte-exact match of the
// `iss` claim in issued tokens and the `issuer` field of the OAuth discovery
// document, so both services read it from the same ISSUER env var.
var Issuer = os.Getenv("ISSUER")

// Comma-separated entity IDs to bootstrap into the Admins group on boot.
// Idempotent; safe to set in any environment.
var AdminEntityIDs = os.Getenv("ADMIN_ENTITY_IDS")

// InternalBootstrapSecret is the shared secret non-core services use to
// exchange for their pre-seeded bearer JWT at startup. Set the same
// value on every service in deploy config; rotating it requires a
// rolling restart so each service can re-fetch its bearer.
//
// The bootstrap endpoint (POST /core/internal/bootstrap-token) refuses
// anything if this is empty — better to fail closed than accept any
// request when the secret hasn't been configured.
var InternalBootstrapSecret = os.Getenv("INTERNAL_BOOTSTRAP_SECRET")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

// ConditionalSyncInterval is how often the periodic conditional-group
// reconcile cron fires. Event-driven sync (member add/remove triggers,
// binding mutations) covers the happy path; the cron is a safety net for
// missed events, batched DB edits, and any future trigger we forget to
// wire. Default 1h; set to 0 (or any non-positive duration) to disable.
var ConditionalSyncInterval = parseDurationOr("CONDITIONAL_SYNC_INTERVAL", time.Hour)

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

var RsaPublicKey *rsa.PublicKey
var RsaPrivateKey *rsa.PrivateKey
var RsaPublicKeyJWKS map[string]interface{}
