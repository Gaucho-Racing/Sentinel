package service

import (
	"sentinel/model"
	"sentinel/utils"
	"time"
)

func GetSubteamsForUser(userID string) []model.Subteam {
	var userSubteams []model.UserSubteam
	result := DB.Where("user_id = ?", userID).Find(&userSubteams)
	if result.Error != nil {
	}
	var subteams []model.Subteam
	for i := range userSubteams {
		subteams = append(subteams, GetSubteamByID(userSubteams[i].RoleID))
	}
	return subteams
}

func SetSubteamsForUser(userID string, subteams []model.UserSubteam) error {
	DB.Where("user_id = ?", userID).Delete(&model.UserSubteam{})
	for _, r := range subteams {
		if result := DB.Create(&r); result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func GetAllSubteams() []model.Subteam {
	var subteams []model.Subteam
	result := DB.Find(&subteams)
	if result.Error != nil {
	}
	return subteams
}

func GetSubteamByID(subteamID string) model.Subteam {
	var subteam model.Subteam
	result := DB.Where("id = ?", subteamID).Find(&subteam)
	if result.Error != nil {
	}
	return subteam
}

func GetSubteamByName(subteamName string) model.Subteam {
	var subteam model.Subteam
	result := DB.Where("name = ?", subteamName).Find(&subteam)
	if result.Error != nil {
	}
	return subteam
}

func CreateSubteam(subteam model.Subteam) error {
	if DB.Where("id = ?", subteam.ID).Updates(&subteam).RowsAffected == 0 {
		utils.SugarLogger.Infoln("New subteam created with id: " + subteam.ID)
		if result := DB.Create(&subteam); result.Error != nil {
			return result.Error
		}
	} else {
		utils.SugarLogger.Infoln("Subteam with id: " + subteam.ID + " has been updated!")
	}
	return nil
}

func InitializeSubteams() {
	CreateSubteam(model.Subteam{
		ID:        "761114473565519882",
		Name:      "Aero",
		CreatedAt: time.Time{},
	})
	CreateSubteam(model.Subteam{
		ID:        "761114557531553824",
		Name:      "Chassis",
		CreatedAt: time.Time{},
	})
	CreateSubteam(model.Subteam{
		ID:        "761114667048763423",
		Name:      "Suspension",
		CreatedAt: time.Time{},
	})
	CreateSubteam(model.Subteam{
		ID:        "761114716985753621",
		Name:      "Powertrain",
		CreatedAt: time.Time{},
	})
	CreateSubteam(model.Subteam{
		ID:        "761116347865890816",
		Name:      "Controls",
		CreatedAt: time.Time{},
	})
	CreateSubteam(model.Subteam{
		ID:        "761331962563919874",
		Name:      "Business",
		CreatedAt: time.Time{},
	})
}
