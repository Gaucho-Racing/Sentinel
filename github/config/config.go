package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const Name = "sentinel-github"
const Version = "5.13.2"
const InternalServiceName = "sentinel-github"

var Env = env("ENV", "PROD")

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

type Configuration struct {
	Port, Org, ClientID, ClientSecret, RedirectURI string
	InstallationID                                 int64
	AppID                                          int64
	PrivateKey                                     []byte
	KerbecsEndpoint, KerbecsUser, KerbecsPassword  string
	BootstrapSecret                                string
	SyncInterval                                   time.Duration
}

func Load() (Configuration, error) {
	interval, err := time.ParseDuration(env("GITHUB_SYNC_INTERVAL", "1h"))
	if err != nil || interval <= 0 {
		return Configuration{}, errors.New("GITHUB_SYNC_INTERVAL must be a positive duration")
	}
	installationID, err := strconv.ParseInt(os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64)
	if err != nil || installationID <= 0 {
		return Configuration{}, errors.New("GITHUB_INSTALLATION_ID must be a positive integer")
	}
	appID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil || appID <= 0 {
		return Configuration{}, errors.New("GITHUB_APP_ID must be a positive integer")
	}
	key := strings.ReplaceAll(os.Getenv("GITHUB_APP_PRIVATE_KEY"), `\n`, "\n")
	if _, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key)); err != nil {
		return Configuration{}, fmt.Errorf("GITHUB_APP_PRIVATE_KEY: %w", err)
	}
	cfg := Configuration{
		Port: env("PORT", "9994"), Org: env("GITHUB_ORG", "gaucho-racing"),
		ClientID: os.Getenv("GITHUB_CLIENT_ID"), ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURI: os.Getenv("GITHUB_REDIRECT_URI"), InstallationID: installationID, AppID: appID,
		PrivateKey: []byte(key), KerbecsEndpoint: env("KERBECS_ENDPOINT", "http://localhost:10300"),
		KerbecsUser: env("KERBECS_USER", "admin"), KerbecsPassword: os.Getenv("KERBECS_PASSWORD"),
		BootstrapSecret: os.Getenv("INTERNAL_BOOTSTRAP_SECRET"), SyncInterval: interval,
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" || cfg.BootstrapSecret == "" || cfg.KerbecsPassword == "" {
		return Configuration{}, errors.New("GitHub client credentials, redirect URI, and bootstrap secret are required")
	}
	if !strings.HasPrefix(cfg.RedirectURI, "https://") && !strings.HasPrefix(cfg.RedirectURI, "http://localhost:") {
		return Configuration{}, errors.New("GITHUB_REDIRECT_URI must use HTTPS or localhost")
	}
	return cfg, nil
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
