package service

import (
	"sentinel/config"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
	"strings"
	"time"
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

func SetRolesForUser(userID string, roles []string) []string {
	existingRoles := GetRolesForUser(userID)
	for _, nr := range roles {
		if !contains(existingRoles, nr) {
			result := database.DB.Create(&model.UserRole{
				UserID:    userID,
				Role:      nr,
				CreatedAt: time.Time{},
			})
			if result.Error != nil {
				utils.SugarLogger.Errorln(result.Error.Error())
			}
		}
	}
	for _, er := range existingRoles {
		if !contains(roles, er) {
			database.DB.Where("user_id = ? AND role = ?", userID, er).Delete(&model.UserRole{})
		}
	}
	return GetRolesForUser(userID)
}

func SyncDiscordRolesForUser(userID string, roleIds []string) {
	subteamRoles := make([]model.UserSubteam, 0)
	roles := GetRolesForUser(userID)
	for _, role := range roles {
		if strings.HasPrefix(role, "d_") {
			roles = removeValue(roles, role)
		}
	}
	for _, id := range roleIds {
		subteam := GetSubteamByID(id)
		if subteam.ID != "" {
			subteamRoles = append(subteamRoles, model.UserSubteam{
				UserID: userID,
				RoleID: subteam.ID,
			})
		} else if id == config.AdminRoleID {
			roles = append(roles, "d_admin")
		} else if id == config.OfficerRoleID {
			roles = append(roles, "d_officer")
		} else if id == config.LeadRoleID {
			roles = append(roles, "d_lead")
		} else if id == config.VerifiedMemberRoleID {
			roles = append(roles, "d_verified")
		} else if id == config.MemberRoleID {
			roles = append(roles, "d_member")
		} else if id == config.AlumniRoleID {
			roles = append(roles, "d_alumni")
		}
	}
	SetSubteamsForUser(userID, subteamRoles)
	SetRolesForUser(userID, roles)

	// Remove all roles except alumni from user
	if contains(roles, "d_alumni") {
		SetRolesForUser(userID, []string{"d_alumni"})
	}

	// If user is not member or alumni, remove all roles
	if !contains(roles, "d_member") && !contains(roles, "d_alumni") {
		SetRolesForUser(userID, []string{})
	}
}

func RemoveAllSubteamDiscordRolesForUser(userID string) {
	subteams := GetSubteamsForUser(userID)
	for _, subteam := range subteams {
		Discord.GuildMemberRoleRemove(config.DiscordGuild, userID, subteam.ID)
	}
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
