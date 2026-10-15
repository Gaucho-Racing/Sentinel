package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
)

const provisioningTriggerTimeout = 30 * time.Second

var provisioningTriggerRoutes = []string{
	"/api/google/reconcile",
	"/api/saml/internal/scim/sync",
}

// TriggerProvisioningSync asks the google and saml services to resync right
// away after an email change or verification, rather than on their next
// scheduled run. Best effort: the schedules still catch anything missed.
func TriggerProvisioningSync(reason string) {
	if config.KerbecsEndpoint == "" {
		return
	}
	go func() {
		token, err := CoreInternalToken()
		if err != nil {
			logger.SugarLogger.Errorf("provisioning trigger (%s): mint token: %v", reason, err)
			return
		}
		for _, route := range provisioningTriggerRoutes {
			if err := postInternal(route, token); err != nil {
				logger.SugarLogger.Errorf("provisioning trigger (%s): %s: %v", reason, route, err)
			}
		}
	}()
}

func postInternal(route string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningTriggerTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(config.KerbecsEndpoint, "/")+route, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
