// Package kerbecs resolves gateway-form paths (e.g. /api/core/entity/1) to the
// concrete upstream URL to call, by asking the kerbecs gateway's admin resolve
// endpoint. It replaces external service-registry route matching: kerbecs is
// already the routing source of truth, so we ask it where a request should go
// and cache the answer locally.
package kerbecs

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const cacheTTL = 5 * time.Minute

var (
	endpoint string
	user     string
	password string
)

type entry struct {
	url string
	exp time.Time
}

var (
	mu    sync.RWMutex
	cache = map[string]entry{}
)

// resolve is a GET, so retries are safe.
var client = resty.New().
	SetTimeout(5 * time.Second).
	SetRetryCount(2).
	SetRetryWaitTime(100 * time.Millisecond).
	AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || (r != nil && r.StatusCode() >= 500)
	})

// Init configures the resolver against the kerbecs admin API. No connection is
// made here — lookups happen lazily on first Resolve — and a background sweeper
// is started to evict expired cache entries.
func Init(adminEndpoint, adminUser, adminPassword string) {
	endpoint = strings.TrimRight(adminEndpoint, "/")
	user = adminUser
	password = adminPassword
	go sweep()
}

type resolveResponse struct {
	Matched       bool   `json:"matched"`
	URL           string `json:"url"`
	RewrittenPath string `json:"rewritten_path"`
}

// Resolve maps a gateway-form path (e.g. /api/core/entity/1) and HTTP method to
// the full upstream URL to call. Answers are cached for cacheTTL.
func Resolve(method, path string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("kerbecs resolver not initialized")
	}
	key := method + " " + path

	mu.RLock()
	if e, ok := cache[key]; ok && time.Now().Before(e.exp) {
		mu.RUnlock()
		return e.url, nil
	}
	mu.RUnlock()

	var rr resolveResponse
	resp, err := client.R().
		SetBasicAuth(user, password).
		SetQueryParam("path", path).
		SetQueryParam("method", method).
		SetResult(&rr).
		Get(endpoint + "/admin-gw/resolve")
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", path, err)
	}
	if resp.StatusCode() == http.StatusNotFound || !rr.Matched {
		return "", fmt.Errorf("no upstream registered for %s", path)
	}
	if resp.IsError() {
		return "", fmt.Errorf("resolve %s: kerbecs returned %d", path, resp.StatusCode())
	}

	full := strings.TrimRight(rr.URL, "/") + rr.RewrittenPath
	mu.Lock()
	cache[key] = entry{url: full, exp: time.Now().Add(cacheTTL)}
	mu.Unlock()
	return full, nil
}

// sweep periodically evicts expired entries so high-cardinality paths (entity
// and token IDs) don't grow the cache without bound.
func sweep() {
	for range time.Tick(cacheTTL) {
		now := time.Now()
		mu.Lock()
		for k, e := range cache {
			if now.After(e.exp) {
				delete(cache, k)
			}
		}
		mu.Unlock()
	}
}
