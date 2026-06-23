package config

import (
	"os"
)

const Name = "sentinel-google"
const Version = "5.6.4"

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

// InternalBootstrapSecret is the shared secret this service uses at
// startup to exchange for its pre-seeded bearer JWT from core. Must
// match core's INTERNAL_BOOTSTRAP_SECRET.
var InternalBootstrapSecret = os.Getenv("INTERNAL_BOOTSTRAP_SECRET")

// InternalServiceName is the SA name on core that this service exchanges
// the bootstrap secret for. Must match a value in
// core/jobs/init.go::InternalServiceAccountNames.
const InternalServiceName = "sentinel-google"

func IsProduction() bool {
	return Env == "PROD"
}
