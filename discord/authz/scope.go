package authz

import "strings"

const (
	SentinelAudience       = "sentinel"
	SentinelAllScope       = "sentinel:all"
	SentinelInternalScope  = "sentinel:internal"
	UserReadScope          = "user:read"
	UserWriteScope         = "user:write"
	GroupsReadScope        = "groups:read"
	GroupsWriteScope       = "groups:write"
	ApplicationsReadScope  = "applications:read"
	ApplicationsWriteScope = "applications:write"
)

func HasScope(scopes string, required string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == required {
			return true
		}
	}
	return false
}

func AudienceContains(audience any, required string) bool {
	switch value := audience.(type) {
	case string:
		return value == required
	case []string:
		for _, candidate := range value {
			if candidate == required {
				return true
			}
		}
	case []any:
		for _, candidate := range value {
			if candidate == required {
				return true
			}
		}
	}
	return false
}

func IsInternalServiceAccount(scopes string, audience any, claims map[string]any) bool {
	if !HasScope(scopes, SentinelInternalScope) || !AudienceContains(audience, SentinelAudience) {
		return false
	}
	accountType, typeOK := claims["type"].(string)
	accountID, idOK := claims["service_account_id"].(string)
	return typeOK && accountType == "service_account" && idOK && accountID != ""
}

func IsFirstPartyUser(scopes string, audience any, claims map[string]any) bool {
	if !HasScope(scopes, SentinelAllScope) || !AudienceContains(audience, SentinelAudience) {
		return false
	}
	entityType, typeOK := claims["entity_type"].(string)
	userID, idOK := claims["user_id"].(string)
	return typeOK && entityType == "USER" && idOK && userID != ""
}
