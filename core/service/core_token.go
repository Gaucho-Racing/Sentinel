package service

import (
	"sync"
	"time"
)

// CoreServiceAccountID is sentinel-core's own service account, seeded at boot.
const CoreServiceAccountID = "sa_01kpy5f8263c4rqnhn9zejxyhk"

const (
	coreTokenLifetime      = time.Hour
	coreTokenRefreshBefore = 5 * time.Minute
)

var coreToken struct {
	sync.Mutex
	value     string
	expiresAt time.Time
}

// CoreServiceToken returns a short-lived bearer for sentinel-core's own
// service account, for calling other Gaucho Racing services such as Notify.
// Core signs it itself, so there is no secret to distribute, and it carries no
// Sentinel scopes, so the receiving service cannot use it against Sentinel.
func CoreServiceToken() (string, error) {
	coreToken.Lock()
	defer coreToken.Unlock()
	if coreToken.value != "" && time.Until(coreToken.expiresAt) > coreTokenRefreshBefore {
		return coreToken.value, nil
	}

	sa, err := GetServiceAccountByID(CoreServiceAccountID)
	if err != nil {
		return "", err
	}
	app, err := GetApplicationByID(sa.ApplicationID)
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(coreTokenLifetime)
	token, _, err := GenerateToken(sa.EntityID, app.ClientID, "", int(coreTokenLifetime.Seconds()), map[string]any{
		"type":               "service_account",
		"service_account_id": sa.ID,
	})
	if err != nil {
		return "", err
	}
	coreToken.value = token
	coreToken.expiresAt = expiresAt
	return token, nil
}
