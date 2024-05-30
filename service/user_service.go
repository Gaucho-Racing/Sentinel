package service

import (
	"fmt"
	"sentinel/model"
	"sentinel/utils"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func GetAllUsers() []model.User {
	var users []model.User
	DB.Find(&users)
	for i := range users {
		users[i].Subteams = GetSubteamsForUser(users[i].ID)
	}
	return users
}

func GetUserByID(userID string) model.User {
	var user model.User
	DB.Where("id = ?", userID).Find(&user)
	user.Subteams = GetSubteamsForUser(user.ID)
	return user
}

func GetUserByUsername(username string) model.User {
	var user model.User
	DB.Where("username = ?", username).Find(&user)
	user.Subteams = GetSubteamsForUser(user.ID)
	return user
}

func GetUserByEmail(email string) model.User {
	var user model.User
	DB.Where("email = ?", email).Find(&user)
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

func SearchUsers(search string) []model.User {
	utils.SugarLogger.Infof("Searching for users with: %s", search)
	var users []model.User
	userStrings := []string{}
	allUsers := GetAllUsers()
	for _, user := range allUsers {
		userStrings = append(userStrings, fmt.Sprintf("%s %s %s %s %s", user.ID, user.Username, user.FirstName, user.LastName, user.Email))
	}
	matches := fuzzy.RankFindNormalizedFold(search, userStrings)
	sort.Sort(matches)
	utils.SugarLogger.Infof("Found %d matches", len(matches))
	for i := 0; i < 5 && i < len(matches); i++ {
		users = append(users, allUsers[matches[i].OriginalIndex])
	}
	return users
}
