package service

import "github.com/gaucho-racing/sentinel/oauth/config"

// IdentityOverride replaces the identity an OAuth client sees so that many real
// Sentinel users can be presented to a relying party as a single shared
// identity (e.g. the shared Google Workspace mailbox team@gauchoracing.com).
// Only the identity claims are swapped — the token `sub` keeps the real entity
// ID, so the authenticated user stays attributable server-side.
type IdentityOverride struct {
	Email     string
	Username  string
	FirstName string
	LastName  string
}

// identityOverrides maps an OAuth client_id to the identity it should assume.
// To add a service: give its Application a stable client_id (set explicitly at
// creation rather than the generated default), surface it as an env-backed
// config value, and register it here. Entries with an empty client_id are
// skipped so an unset env var can never match clientID == "".
var identityOverrides = buildIdentityOverrides()

func buildIdentityOverrides() map[string]IdentityOverride {
	overrides := map[string]IdentityOverride{}
	register := func(clientID string, o IdentityOverride) {
		if clientID == "" {
			return
		}
		overrides[clientID] = o
	}

	register(config.TeamGoogleClientID, IdentityOverride{
		Email:     "team@gauchoracing.com",
		Username:  "team",
		FirstName: "Gaucho",
		LastName:  "Racing",
	})

	return overrides
}

// applyIdentityOverride returns e with its identity fields replaced by the
// override registered for clientID, if any. The entity ID is left untouched so
// the issued token's `sub` still identifies the real user. Returns e unchanged
// when no override applies.
func applyIdentityOverride(clientID string, e oidcEntity) oidcEntity {
	o, ok := identityOverrides[clientID]
	if !ok {
		return e
	}
	e.EmailAuth.Email = o.Email
	if e.User == nil {
		e.User = &oidcUser{}
	}
	e.User.Email = o.Email
	e.User.Username = o.Username
	e.User.FirstName = o.FirstName
	e.User.LastName = o.LastName
	e.User.AvatarURL = ""
	return e
}
