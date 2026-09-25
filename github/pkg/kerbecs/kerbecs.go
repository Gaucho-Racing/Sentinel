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
	mu       sync.RWMutex
	cache    = map[string]entry{}
)

type entry struct {
	url string
	exp time.Time
}

var client = resty.New().SetTimeout(5 * time.Second).SetRetryCount(2).SetRetryWaitTime(100 * time.Millisecond).
	AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || (r != nil && r.StatusCode() >= 500)
	})

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
	resp, err := client.R().SetBasicAuth(user, password).SetQueryParam("path", path).
		SetQueryParam("method", method).SetResult(&rr).Get(endpoint + "/admin-gw/resolve")
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

func sweep() {
	for range time.Tick(cacheTTL) {
		now := time.Now()
		mu.Lock()
		for key, value := range cache {
			if now.After(value.exp) {
				delete(cache, key)
			}
		}
		mu.Unlock()
	}
}
