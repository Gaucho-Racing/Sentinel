package config

import (
	"crypto/rsa"
	"os"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "Sentinel Core",
	Version:     "5.2.0",
	Endpoint:    os.Getenv("SERVICE_ENDPOINT"),
	HealthCheck: os.Getenv("SERVICE_HEALTH_CHECK"),
}

// The /** glob in rincon-go matches a non-empty suffix, so a bare path like
// /groups or /users won't match /groups/** — register both forms so the
// collection-level GETs are reachable via service-to-service calls.
var Routes = []rincon.Route{
	{Route: "/core/**", Method: "*"},
	{Route: "/users", Method: "*"},
	{Route: "/users/**", Method: "*"},
	{Route: "/applications", Method: "*"},
	{Route: "/applications/**", Method: "*"},
	{Route: "/groups", Method: "*"},
	{Route: "/groups/**", Method: "*"},
	{Route: "/entities", Method: "*"},
	{Route: "/entities/**", Method: "*"},
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

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
