package config

import (
	"crypto/rsa"
	"os"

	"github.com/bk1031/rincon-go/v2"
)

var Service rincon.Service = rincon.Service{
	Name:        "sentinel-core",
	Version:     "0.1.0",
	Endpoint:    os.Getenv("SERVICE_ENDPOINT"),
	HealthCheck: os.Getenv("SERVICE_HEALTH_CHECK"),
}

var Routes = []rincon.Route{
	{
		Route:  "/core/ping",
		Method: "GET",
	},
	{
		Route:  "/core/**",
		Method: "*",
	},
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

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
