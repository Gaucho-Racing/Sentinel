package service

import (
	"sentinel/config"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
	"strings"
)

func GetSubteamsForUser(userID string) []model.Subteam {
	var userSubteams []model.UserSubteam
	database.DB.Where("user_id = ?", userID).Find(&userSubteams)
	var subteams []model.Subteam
	for i := range userSubteams {
		subteams = append(subteams, GetSubteamByID(userSubteams[i].RoleID))
	}
	return subteams
}

func SetSubteamsForUser(userID string, subteams []model.UserSubteam) error {
	database.DB.Where("user_id = ?", userID).Delete(&model.UserSubteam{})
	for _, r := range subteams {
		if result := database.DB.Create(&r); result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func GetAllSubteams() []model.Subteam {
	var subteams []model.Subteam
	database.DB.Find(&subteams)
	return subteams
}

func GetSubteamByID(subteamID string) model.Subteam {
	var subteam model.Subteam
	database.DB.Where("id = ?", subteamID).Find(&subteam)
	return subteam
}

func GetSubteamByName(subteamName string) model.Subteam {
	var subteam model.Subteam
	database.DB.Where("name = ?", subteamName).Find(&subteam)
	return subteam
}

func CreateSubteam(subteam model.Subteam) error {
	if database.DB.Where("id = ?", subteam.ID).Updates(&subteam).RowsAffected == 0 {
		utils.SugarLogger.Infoln("New subteam created with id: " + subteam.ID)
		if result := database.DB.Create(&subteam); result.Error != nil {
			return result.Error
		}
	} else {
		utils.SugarLogger.Infoln("Subteam with id: " + subteam.ID + " has been updated!")
	}
	return nil
}

func DeleteAllSubteams() error {
	if result := database.DB.Where("1 = 1").Delete(&model.Subteam{}); result.Error != nil {
		return result.Error
	}
	return nil
}

func InitializeSubteams() {
	g, err := Discord.Guild(config.DiscordGuild)
	if err != nil {
		utils.SugarLogger.Errorln("Error getting guild,", err)
		return
	}
	DeleteAllSubteams()
	for _, r := range g.Roles {
		for _, name := range config.SubteamRoleNames {
			if strings.Contains(strings.ToLower(r.Name), strings.ToLower(name)) {
				utils.SugarLogger.Infof("Found subteam role: %s for %s", r.ID, name)
				CreateSubteam(model.Subteam{
					ID:   r.ID,
					Name: name,
				})
			}
		}
	}
}
