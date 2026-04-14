package model

var ValidScopes = map[string]string{
	"user:read":          "Read user and entity profile information",
	"user:write":         "Update user profile information",
	"groups:read":        "Read group memberships",
	"applications:read":  "Read application details",
	"applications:write": "Manage applications",
	"sentinel:all":       "Full internal access (not available to third-party apps)",
}
