package service

import (
	"sync"
	"time"

	"github.com/gaucho-racing/sentinel/core/authz"
)

const CoreServiceAccountID = "sa_01kpy5f8263c4rqnhn9zejxyhk"

const (
	coreTokenLifetime      = time.Hour
	coreTokenRefreshBefore = 5 * time.Minute
)

type cachedCoreToken struct {
	value     string
	expiresAt time.Time
}

var coreTokens = struct {
	sync.Mutex
	byScope map[string]cachedCoreToken
}{byScope: map[string]cachedCoreToken{}}

// CoreServiceToken is a scopeless token for sentinel-core's own service
// account, for services like Notify that only need to know who is calling.
func CoreServiceToken() (string, error) {
	return coreToken("")
}

// CoreInternalToken carries sentinel:internal, for calling Sentinel's own
// services such as google and saml.
func CoreInternalToken() (string, error) {
	return coreToken(authz.SentinelInternalScope)
}

func coreToken(scope string) (string, error) {
	coreTokens.Lock()
	defer coreTokens.Unlock()
	if cached, ok := coreTokens.byScope[scope]; ok && time.Until(cached.expiresAt) > coreTokenRefreshBefore {
		return cached.value, nil
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
	token, _, err := GenerateToken(sa.EntityID, app.ClientID, scope, int(coreTokenLifetime.Seconds()), map[string]any{
		"type":               "service_account",
		"service_account_id": sa.ID,
	})
	if err != nil {
		return "", err
	}
	coreTokens.byScope[scope] = cachedCoreToken{value: token, expiresAt: expiresAt}
	return token, nil
}
