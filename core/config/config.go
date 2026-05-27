package config

import (
	"crypto/rsa"
	"os"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "Sentinel Core",
	Version:     "5.0.0",
	Endpoint:    os.Getenv("SERVICE_ENDPOINT"),
	HealthCheck: os.Getenv("SERVICE_HEALTH_CHECK"),
}

var Routes = []rincon.Route{
	{
		Route:  "/core/**",
		Method: "*",
	},
	{
		Route:  "/users/**",
		Method: "*",
	},
	{
		Route:  "/applications/**",
		Method: "*",
	},
	{
		Route:  "/groups/**",
		Method: "*",
	},
	{
		Route:  "/entities/**",
		Method: "*",
	},
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

// Optional dev convenience: when set, the seed promotes this entity to owner
// of the seeded test group so the developer can act on the inbox UI.
var DevSeedOwnerEntityID = os.Getenv("DEV_SEED_OWNER_ENTITY_ID")

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
