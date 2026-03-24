package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		return []model.User{}, err
	}
	for i := range users {
		PopulateUser(&users[i])
	}
	return users, nil
}

func GetUserByID(id string) (model.User, error) {
	var user model.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return model.User{}, err
	}
	PopulateUser(&user)
	return user, nil
}

func GetUserByEntityID(entityID string) (model.User, error) {
	var user model.User
	if err := database.DB.Where("entity_id = ?", entityID).First(&user).Error; err != nil {
		return model.User{}, err
	}
	PopulateUser(&user)
	return user, nil
}

func CreateUser(user model.User) (model.User, error) {
	if user.ID == "" {
		user.ID = ulid.Make().Prefixed("usr")
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return model.User{}, err
	}
	PopulateUser(&user)
	return user, nil
}

func UpdateUser(user model.User) (model.User, error) {
	if err := database.DB.Save(&user).Error; err != nil {
		return model.User{}, err
	}
	PopulateUser(&user)
	return user, nil
}

func DeleteUser(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		return err
	}
	return nil
}

func PopulateUser(user *model.User) {
	groups, err := GetGroupsForEntity(user.EntityID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get groups for user %s: %v", user.ID, err)
	}
	user.Groups = groups

	emailAuth, err := GetEmailAuthForEntity(user.EntityID)
	if err == nil {
		user.Email = emailAuth.Email
	}

	phoneAuth, err := GetPhoneAuthForEntity(user.EntityID)
	if err == nil {
		user.PhoneNumber = phoneAuth.PhoneNumber
	}
}

func GetGroupsForEntity(entityID string) ([]model.Group, error) {
	var members []model.GroupMember
	if err := database.DB.Where("entity_id = ?", entityID).Find(&members).Error; err != nil {
		return []model.Group{}, err
	}
	var groups []model.Group
	for _, member := range members {
		var group model.Group
		if err := database.DB.Where("id = ?", member.GroupID).First(&group).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to get group %s for entity %s: %v", member.GroupID, entityID, err)
			continue
		}
		groups = append(groups, group)
	}
	return groups, nil
}
