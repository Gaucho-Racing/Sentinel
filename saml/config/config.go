package config

import (
	"os"
)

const Name = "sentinel-saml"
const Version = "5.5.2"

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

// Issuer is the public base URL of the deployment. It also serves as the SAML
// IdP entityID (the metadata URL is derived from it), so it must be stable —
// changing it invalidates trust with every registered SP.
var Issuer = os.Getenv("ISSUER")

// SentinelClientID is the first-party identifier whose application row carries
// the group links that act as a global default added to every app's filter
// set — same semantics as in the oauth service.
const SentinelClientID = "sentinel"

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

func IsProduction() bool {
	return Env == "PROD"
}

// MetadataPath / SSOPath are the IdP's public endpoints, served at the issuer
// root (no /api prefix) so the URLs published in metadata are clean and stable.
const MetadataPath = "/saml/metadata"
const SSOPath = "/saml/sso"

// AuthorizePath is the SPA consent route the SSO endpoint redirects the browser
// to. The SPA holds the first-party session and posts the approved entity back.
const AuthorizePath = "/saml/authorize"
