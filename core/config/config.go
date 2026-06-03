package config

import (
	"crypto/rsa"
	"os"
)

const Name = "sentinel-core"
const Version = "5.5.1"

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

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

func IsProduction() bool {
	return Env == "PROD"
}

var RsaPublicKey *rsa.PublicKey
var RsaPrivateKey *rsa.PrivateKey
var RsaPublicKeyJWKS map[string]interface{}
