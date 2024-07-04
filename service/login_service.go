package service

import (
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"

	"github.com/google/uuid"
)

func GetAllLogins() []model.UserLogin {
	var logins []model.UserLogin
	database.DB.Order("created_at DESC").Find(&logins)
	return logins
}

func GetLoginsForUser(userID string) []model.UserLogin {
	var logins []model.UserLogin
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&logins)
	return logins
}

func GetLastNLoginsForUser(userID string, n int) []model.UserLogin {
	var logins []model.UserLogin
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(n).Find(&logins)
	return logins
}

func GetLoginsForDestination(destination string) []model.UserLogin {
	var logins []model.UserLogin
	database.DB.Where("destination = ?", destination).Order("created_at DESC").Find(&logins)
	return logins
}

func GetLastNLoginsForDestination(destination string, n int) []model.UserLogin {
	var logins []model.UserLogin
	database.DB.Where("destination = ?", destination).Order("created_at DESC").Limit(n).Find(&logins)
	return logins
}

func GetLastLoginForUserToDestinationWithScopes(userID string, destination string, scope string) model.UserLogin {
	var login model.UserLogin
	database.DB.Where("user_id = ? AND destination = ? AND scope = ?", userID, destination, scope).Order("created_at DESC").First(&login)
	return login
}

func GetLoginByID(loginID string) model.UserLogin {
	var login model.UserLogin
	database.DB.Where("id = ?", loginID).First(&login)
	return login
}

func CreateLogin(login model.UserLogin) error {
	if login.ID == "" {
		login.ID = uuid.New().String()
	}
	if database.DB.Where("id = ?", login.ID).Updates(&login).RowsAffected == 0 {
		utils.SugarLogger.Infof("New login from %s for %s", login.UserID, login.Destination)
		if result := database.DB.Create(&login); result.Error != nil {
			return result.Error
		}
	}
	return nil
}
