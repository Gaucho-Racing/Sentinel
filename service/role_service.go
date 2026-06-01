package service

import (
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
)

func GetRolesForUser(userID string) []string {
	var roles []model.UserRole
	var roleNames = make([]string, 0)
	result := database.DB.Where("user_id = ?", userID).Find(&roles)
	if result.Error != nil {
		return roleNames
	}
	for _, r := range roles {
		roleNames = append(roleNames, r.Role)
	}
	return roleNames
}

func GetDiscordRolesForUser(userID string) []string {
	var roles []model.UserRole
	var roleNames = make([]string, 0)
	result := database.DB.Where("user_id = ? AND role LIKE ?", userID, "d_%").Find(&roles)
	if result.Error != nil {
		return roleNames
	}
	for _, r := range roles {
		roleNames = append(roleNames, r.Role)
	}
	return roleNames
}

// SetRolesForUser is a no-op. Role syncing is now owned by Sentinel v5.
// Returns the current roles unchanged.
func SetRolesForUser(userID string, roles []string) []string {
	utils.SugarLogger.Infof("SetRolesForUser(%s) called but role syncing is disabled in v4 (owned by v5)", userID)
	return GetRolesForUser(userID)
}

// SyncDiscordRolesForUser is a no-op. Role syncing is now owned by Sentinel v5.
func SyncDiscordRolesForUser(userID string, roleIds []string) {
	utils.SugarLogger.Infof("SyncDiscordRolesForUser(%s) called but role syncing is disabled in v4 (owned by v5)", userID)
}

// RemoveAllSubteamDiscordRolesForUser is a no-op. Role syncing is now owned by Sentinel v5.
func RemoveAllSubteamDiscordRolesForUser(userID string) {
	utils.SugarLogger.Infof("RemoveAllSubteamDiscordRolesForUser(%s) called but role syncing is disabled in v4 (owned by v5)", userID)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func removeValue(s []string, value string) []string {
	result := []string{}
	for _, v := range s {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
