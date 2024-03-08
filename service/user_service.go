package service

import (
	"sentinel/model"
	"sentinel/utils"
)

func GetAllUsers() []model.User {
	var users []model.User
	result := DB.Find(&users)
	if result.Error != nil {
	}
	for i := range users {
		users[i].Subteams = GetSubteamsForUser(users[i].ID)
	}
	return users
}

func GetUserByID(userID string) model.User {
	var user model.User
	result := DB.Where("id = ?", userID).Find(&user)
	if result.Error != nil {
	}
	user.Subteams = GetSubteamsForUser(user.ID)
	return user
}

func CreateUser(user model.User) error {
	if DB.Where("id = ?", user.ID).Updates(&user).RowsAffected == 0 {
		utils.SugarLogger.Infoln("New user created with id: " + user.ID)
		if result := DB.Create(&user); result.Error != nil {
			return result.Error
		}
		go DiscordLogNewUser(user)
	} else {
		utils.SugarLogger.Infoln("User with id: " + user.ID + " has been updated!")
	}
	return nil
}

func DeleteUser(userID string) error {
	if result := DB.Where("id = ?", userID).Delete(&model.User{}); result.Error != nil {
		return result.Error
	}
	return nil
}
