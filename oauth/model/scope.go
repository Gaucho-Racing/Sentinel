package model

var ValidScopes = map[string]string{
	"openid":             "Authenticate you and issue an ID token",
	"profile":            "Read your basic profile (name, username, picture)",
	"email":              "Read your email address",
	"offline_access":     "Stay signed in without re-authenticating (refresh token)",
	"user:read":          "Read user and entity profile information",
	"user:write":         "Update user profile information",
	"groups:read":        "Read group memberships",
	"applications:read":  "Read application details",
	"applications:write": "Manage applications",
	"sentinel:all":       "Full internal access (not available to third-party apps)",
}
