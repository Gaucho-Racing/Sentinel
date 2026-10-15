package service

import (
	"sync"
	"time"
)

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
