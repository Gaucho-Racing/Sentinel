package utils

import "sentinel/config"

// IsAdmin checks if the user has the admin role.
// The function takes in a list of role IDs and returns a boolean.
func IsAdmin(roles []string) bool {
	for _, role := range roles {
		if role == config.AdminRoleID {
			return true
		}
	}
	return false
}

// IsOfficer checks if the user has the officer role.
// The function takes in a list of role IDs and returns a boolean.
func IsOfficer(roles []string) bool {
	for _, role := range roles {
		if role == config.OfficerRoleID {
			return true
		}
	}
	return false
}

// IsLead checks if the user has the lead role.
// The function takes in a list of role IDs and returns a boolean.
func IsLead(roles []string) bool {
	for _, role := range roles {
		if role == config.LeadRoleID {
			return true
		}
	}
	return false
}

// isInnerCircle checks if the user has any of the inner circle roles.
// The function takes in a list of role IDs and returns a boolean.
func IsInnerCircle(roles []string) bool {
	return IsAdmin(roles) || IsOfficer(roles) || IsLead(roles)
}
