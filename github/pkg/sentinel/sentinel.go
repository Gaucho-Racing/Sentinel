package sentinel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gaucho-racing/sentinel/github/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/github/pkg/logger"
	"github.com/go-resty/resty/v2"
)

var (
	bearer   string
	bearerMu sync.RWMutex
)

func SetBearer(token string) {
	bearerMu.Lock()
	defer bearerMu.Unlock()
	bearer = token
}

func getBearer() string {
	bearerMu.RLock()
	defer bearerMu.RUnlock()
	return bearer
}

func Bootstrap(serviceName, secret string) error {
	if secret == "" {
		return errors.New("INTERNAL_BOOTSTRAP_SECRET is not configured")
	}
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		var result struct {
			Token string `json:"token"`
		}
		err := Post(context.Background(), "/api/core/internal/bootstrap-token", map[string]string{"name": serviceName}, &result, map[string]string{"X-Bootstrap-Secret": secret})
		if err == nil && result.Token != "" {
			SetBearer(result.Token)
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New("bootstrap exchange returned empty token")
		}
		logger.SugarLogger.Warnf("bootstrap attempt %d failed: %v", attempt+1, lastErr)
	}
	return fmt.Errorf("bootstrap failed after retries: %w", lastErr)
}

var ErrRouteResolution = errors.New("could not resolve route via kerbecs")

var client = resty.New().SetTimeout(5 * time.Second).SetRetryCount(2).SetRetryWaitTime(100 * time.Millisecond).
	AddRetryCondition(func(r *resty.Response, err error) bool {
		if r == nil || r.Request == nil || r.Request.Method != http.MethodGet {
			return false
		}
		return err != nil || r.StatusCode() >= 500
	})

type APIError struct {
	Method  string
	Route   string
	Status  int
	Body    string
	Message string
	Err     error
}

func (e *APIError) Error() string {
	if e.Status == 0 {
		return fmt.Sprintf("%s %s: %v", e.Method, e.Route, e.Err)
	}
	if e.Message != "" {
		return fmt.Sprintf("%s %s returned %d: %s", e.Method, e.Route, e.Status, e.Message)
	}
	return fmt.Sprintf("%s %s returned %d", e.Method, e.Route, e.Status)
}

func (e *APIError) Unwrap() error { return e.Err }

func do(ctx context.Context, method, route string, body, result interface{}, headers []map[string]string) error {
	url, err := kerbecs.Resolve(method, route)
	if err != nil {
		return &APIError{Method: method, Route: route, Err: fmt.Errorf("%w: %v", ErrRouteResolution, err)}
	}
	req := client.R().SetContext(ctx)
	explicitBearer := len(headers) > 0 && headers[0]["Authorization"] != ""
	if token := getBearer(); token != "" && !explicitBearer {
		req.SetAuthToken(token)
	}
	if body != nil {
		req.SetBody(body)
	}
	if result != nil {
		req.SetResult(result)
	}
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Execute(method, url)
	if err != nil {
		return &APIError{Method: method, Route: route, Err: err}
	}
	if resp.IsError() {
		ae := &APIError{Method: method, Route: route, Status: resp.StatusCode(), Body: resp.String()}
		var parsed struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(resp.Body(), &parsed) == nil {
			ae.Message = parsed.Error
		}
		logger.SugarLogger.Errorf("%s %s returned %d: %s", method, route, resp.StatusCode(), resp.String())
		return ae
	}
	return nil
}

func Get(ctx context.Context, route string, result interface{}, headers ...map[string]string) error {
	return do(ctx, http.MethodGet, route, nil, result, headers)
}
func Post(ctx context.Context, route string, body, result interface{}, headers ...map[string]string) error {
	return do(ctx, http.MethodPost, route, body, result, headers)
}
func Put(ctx context.Context, route string, body, result interface{}, headers ...map[string]string) error {
	return do(ctx, http.MethodPut, route, body, result, headers)
}
func Delete(ctx context.Context, route string, result interface{}, headers ...map[string]string) error {
	return do(ctx, http.MethodDelete, route, nil, result, headers)
}
