package service

import (
	"regexp"
	"strings"
	"sync"
)

// MatchRedirectURI checks whether a concrete redirect URI matches a registered
// pattern. Patterns support `*` as a wildcard for any sequence of characters,
// in any position (scheme, host, port, path, query). An exact pattern with no
// `*` matches only itself.
//
// Security note: wildcard redirect URIs widen the attack surface (open
// redirect, subdomain takeover, path traversal). The matcher is intentionally
// strict outside of `*` — every other character must match literally — so the
// only way to broaden a registration is by adding an explicit `*`.
func MatchRedirectURI(pattern, uri string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == uri
	}
	re, err := compileRedirectPattern(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(uri)
}

var (
	redirectPatternCache   = map[string]*regexp.Regexp{}
	redirectPatternCacheMu sync.RWMutex
)

func compileRedirectPattern(pattern string) (*regexp.Regexp, error) {
	redirectPatternCacheMu.RLock()
	if re, ok := redirectPatternCache[pattern]; ok {
		redirectPatternCacheMu.RUnlock()
		return re, nil
	}
	redirectPatternCacheMu.RUnlock()

	// Escape regex metacharacters, then turn the (now-escaped) `*` back into
	// `.*` so it acts as a wildcard. Anchored so partial matches don't slip
	// through.
	escaped := regexp.QuoteMeta(pattern)
	expr := "^" + strings.ReplaceAll(escaped, `\*`, ".*") + "$"
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	redirectPatternCacheMu.Lock()
	redirectPatternCache[pattern] = re
	redirectPatternCacheMu.Unlock()
	return re, nil
}
