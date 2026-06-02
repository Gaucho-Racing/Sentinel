package sentinel

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/pkg/rincon"
	"github.com/go-resty/resty/v2"
)

// A short per-request timeout plus a couple of retries softens transient blips
// on service-to-service calls. Retries are limited to idempotent GETs —
// retrying a POST could double-apply a non-idempotent write.
var client = resty.New().
	SetTimeout(5 * time.Second).
	SetRetryCount(2).
	SetRetryWaitTime(100 * time.Millisecond).
	AddRetryCondition(func(r *resty.Response, err error) bool {
		if r == nil || r.Request == nil || r.Request.Method != http.MethodGet {
			return false
		}
		return err != nil || r.StatusCode() >= 500
	})

func resolveURL(route string, method string) (string, error) {
	if rincon.RinconClient == nil {
		return "", fmt.Errorf("rincon client is not initialized")
	}
	service, err := rincon.RinconClient.MatchRoute(route, method)
	if err != nil {
		return "", fmt.Errorf("failed to resolve route %s: %w", route, err)
	}
	return service.Endpoint + route, nil
}

func Get(route string, result interface{}, headers ...map[string]string) error {
	url, err := resolveURL(route, "GET")
	if err != nil {
		return err
	}
	req := client.R().SetResult(result)
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Get(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.SugarLogger.Errorf("GET %s returned %d: %s", route, resp.StatusCode(), resp.String())
		return fmt.Errorf("GET %s returned %d", route, resp.StatusCode())
	}
	return nil
}

func Post(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	url, err := resolveURL(route, "POST")
	if err != nil {
		return err
	}
	req := client.R().SetBody(body).SetResult(result)
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Post(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.SugarLogger.Errorf("POST %s returned %d: %s", route, resp.StatusCode(), resp.String())
		return fmt.Errorf("POST %s returned %d", route, resp.StatusCode())
	}
	return nil
}

func Put(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	url, err := resolveURL(route, "PUT")
	if err != nil {
		return err
	}
	req := client.R().SetBody(body).SetResult(result)
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Put(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.SugarLogger.Errorf("PUT %s returned %d: %s", route, resp.StatusCode(), resp.String())
		return fmt.Errorf("PUT %s returned %d", route, resp.StatusCode())
	}
	return nil
}

func Patch(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	url, err := resolveURL(route, "PATCH")
	if err != nil {
		return err
	}
	req := client.R().SetBody(body).SetResult(result)
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Patch(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.SugarLogger.Errorf("PATCH %s returned %d: %s", route, resp.StatusCode(), resp.String())
		return fmt.Errorf("PATCH %s returned %d", route, resp.StatusCode())
	}
	return nil
}

func Delete(route string, result interface{}, headers ...map[string]string) error {
	url, err := resolveURL(route, "DELETE")
	if err != nil {
		return err
	}
	req := client.R().SetResult(result)
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Delete(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.SugarLogger.Errorf("DELETE %s returned %d: %s", route, resp.StatusCode(), resp.String())
		return fmt.Errorf("DELETE %s returned %d", route, resp.StatusCode())
	}
	return nil
}
