package utils

import "sentinel/config"

func IsAdmin(roles []string) bool {
	for _, role := range roles {
		if role == config.AdminRoleID {
			return true
		}
	}
	return false
}

func IsOfficer(roles []string) bool {
	for _, role := range roles {
		if role == config.OfficerRoleID {
			return true
		}
	}
	return false
}

func IsLead(roles []string) bool {
	for _, role := range roles {
		if role == config.LeadRoleID {
			return true
		}
	}
	return false
}
