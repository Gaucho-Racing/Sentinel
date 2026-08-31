package api

import "testing"

func TestParseRefreshTokenClaimsRequiresExactAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience interface{}
		valid    bool
	}{
		{name: "string", audience: "sentinel", valid: true},
		{name: "single item array", audience: []interface{}{"sentinel"}, valid: true},
		{name: "wrong client", audience: []interface{}{"third-party"}},
		{name: "multiple audiences", audience: []interface{}{"sentinel", "third-party"}},
		{name: "missing audience"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := map[string]interface{}{
				"sub":   "ent_1",
				"scope": "openid refresh_token",
				"jti":   "jwt_1",
				"aud":   test.audience,
			}
			parsed, err := parseRefreshTokenClaims(claims, "sentinel")
			if test.valid && err != nil {
				t.Fatalf("expected valid claims, got %v", err)
			}
			if !test.valid && err == nil {
				t.Fatalf("expected invalid claims, got %#v", parsed)
			}
		})
	}
}

func TestParseRefreshTokenClaimsRequiresRefreshScopeAndTokenID(t *testing.T) {
	tests := []map[string]interface{}{
		{"sub": "ent_1", "scope": "openid", "jti": "jwt_1", "aud": "sentinel"},
		{"sub": "ent_1", "scope": "openid refresh_token", "aud": "sentinel"},
		{"scope": "openid refresh_token", "jti": "jwt_1", "aud": "sentinel"},
	}

	for _, claims := range tests {
		if parsed, err := parseRefreshTokenClaims(claims, "sentinel"); err == nil {
			t.Fatalf("expected invalid claims, got %#v", parsed)
		}
	}
}
