package model

import "github.com/gaucho-racing/sentinel/oauth/authz"

var ValidScopes = map[string]string{
	"openid":                     "Authenticate you and issue an ID token",
	"profile":                    "Read your basic profile (name, username, picture)",
	"email":                      "Read your email address",
	"offline_access":             "Stay signed in without re-authenticating (refresh token)",
	authz.UserReadScope:          "Read user and entity profile information",
	authz.UserWriteScope:         "Update user profile information",
	authz.GroupsReadScope:        "Read group memberships",
	authz.GroupsWriteScope:       "Manage groups and group memberships",
	authz.ApplicationsReadScope:  "Read application details",
	authz.ApplicationsWriteScope: "Manage applications",
	authz.SentinelAllScope:       "Full read and write access to Sentinel resources",
	authz.SentinelInternalScope:  "Full internal service access",
}
