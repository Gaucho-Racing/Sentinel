package sentinel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gaucho-racing/sentinel/google/pkg/kerbecs"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
	"github.com/go-resty/resty/v2"
)

// bearer is the service-account JWT this process uses to authenticate
// to other Sentinel services. Configured once at startup via Bootstrap.
// Reads via RLock so the request hot path doesn't serialize on writes.
var (
	bearer   string
	bearerMu sync.RWMutex
)

// SetBearer wires a bearer token into the client. Subsequent Get/Post/...
// calls send it as Authorization: Bearer. Empty string clears the
// header — useful for tests that want to exercise the unauth'd path.
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

// Bootstrap exchanges INTERNAL_BOOTSTRAP_SECRET for this service's
// pre-seeded bearer JWT and configures the client. Call once at
// startup, before any other sentinel request. The bootstrap call
// itself goes out without a bearer; core's /core/internal/bootstrap-token
// validates the shared secret in the X-Bootstrap-Secret header instead.
//
// Retries with linear backoff (~10s total) to absorb the docker-compose
// boot race — sentinel-core's HTTP listener may not be up the moment
// this service starts, even with depends_on.
func Bootstrap(serviceName, secret string) error {
	if secret == "" {
		return errors.New("INTERNAL_BOOTSTRAP_SECRET is not configured")
	}
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		var out struct {
			Token string `json:"token"`
		}
		err := Post(
			"/api/core/internal/bootstrap-token",
			map[string]string{"name": serviceName},
			&out,
			map[string]string{"X-Bootstrap-Secret": secret},
		)
		if err == nil && out.Token != "" {
			SetBearer(out.Token)
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

// Sentinel-side error categories — wrapped into APIError.Err so callers
// can errors.Is on them and pick the right user-facing message.
var (
	ErrRouteResolution = errors.New("could not resolve route via kerbecs")
)

// A short per-request timeout plus a couple of retries softens transient core
// blips so authz-relevant reads (group links, entity groups) don't fail closed
// over a momentary hiccup. Retries are limited to idempotent GETs — retrying a
// POST (token mint, login record) could double-issue.
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

// APIError is returned by every method in this package. Status == 0 means no
// HTTP response was received (route resolution failure or transport error).
// Status > 0 means the upstream replied with that status code. Callers should
// use errors.As to inspect it and decide how to surface to their own
// response — most importantly, a 4xx from upstream should NOT collapse to a
// generic "service unavailable" on the user-facing side.
type APIError struct {
	Method  string
	Route   string
	Status  int    // 0 when no HTTP response was received
	Body    string // raw response body
	Message string // parsed "error" field from a JSON body, when present
	Err     error  // underlying transport or resolution error
}

func (e *APIError) Error() string {
	if e.Status == 0 {
		if e.Err != nil {
			return fmt.Sprintf("%s %s: %v", e.Method, e.Route, e.Err)
		}
		return fmt.Sprintf("%s %s: no response", e.Method, e.Route)
	}
	if e.Message != "" {
		return fmt.Sprintf("%s %s returned %d: %s", e.Method, e.Route, e.Status, e.Message)
	}
	return fmt.Sprintf("%s %s returned %d", e.Method, e.Route, e.Status)
}

func (e *APIError) Unwrap() error { return e.Err }

func resolveURL(route string, method string) (string, error) {
	url, err := kerbecs.Resolve(method, route)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrRouteResolution, route, err)
	}
	return url, nil
}

// do executes the request and converts any failure path into an *APIError.
// success returns nil; resty unmarshals the response body into result for us.
func do(method, route string, body, result interface{}, headers []map[string]string) error {
	url, err := resolveURL(route, method)
	if err != nil {
		return &APIError{Method: method, Route: route, Err: err}
	}
	req := client.R()
	// Attach the service's bearer when one is set — Bootstrap installs
	// it at startup. The explicit `headers` param (used by Bootstrap
	// itself for the X-Bootstrap-Secret header) is additive, applied
	// after SetAuthToken.
	if b := getBearer(); b != "" {
		req = req.SetAuthToken(b)
	}
	if body != nil {
		req = req.SetBody(body)
	}
	if result != nil {
		req = req.SetResult(result)
	}
	if len(headers) > 0 {
		req = req.SetHeaders(headers[0])
	}
	resp, err := req.Execute(method, url)
	if err != nil {
		return &APIError{Method: method, Route: route, Err: err}
	}
	if resp.IsError() {
		ae := &APIError{
			Method: method,
			Route:  route,
			Status: resp.StatusCode(),
			Body:   resp.String(),
		}
		var parsed struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(resp.Body(), &parsed) == nil && parsed.Error != "" {
			ae.Message = parsed.Error
		}
		logger.SugarLogger.Errorf("%s %s returned %d: %s", method, route, resp.StatusCode(), resp.String())
		return ae
	}
	return nil
}

func Get(route string, result interface{}, headers ...map[string]string) error {
	return do("GET", route, nil, result, headers)
}

func Post(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	return do("POST", route, body, result, headers)
}

func Put(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	return do("PUT", route, body, result, headers)
}

func Patch(route string, body interface{}, result interface{}, headers ...map[string]string) error {
	return do("PATCH", route, body, result, headers)
}

func Delete(route string, result interface{}, headers ...map[string]string) error {
	return do("DELETE", route, nil, result, headers)
}
