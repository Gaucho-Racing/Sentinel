package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllUsers() ([]model.User, error) {
	users := []model.User{}
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

// IsUsernameAvailable returns true when no user has the given username,
// matched case-insensitively.
func IsUsernameAvailable(username string) (bool, error) {
	var count int64
	if err := database.DB.Model(&model.User{}).
		Where("LOWER(username) = LOWER(?)", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
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
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	user.Groups = names

	emailAuth, err := GetEmailAuthForEntity(user.EntityID)
	if err == nil {
		user.Email = emailAuth.Email
	}

	phoneAuth, err := GetPhoneAuthForEntity(user.EntityID)
	if err == nil {
		user.PhoneNumber = phoneAuth.PhoneNumber
	}
}

// GetGroupsForEntity loads an entity's groups in two queries — the memberships,
// then every referenced group at once. Fetching one group per membership made
// this cost a round trip per group, and since it runs on the token-issuing path
// (several times per login) it made auth latency scale with a user's group
// count, to the point of exceeding the timeouts relying parties allow.
func GetGroupsForEntity(entityID string) ([]model.Group, error) {
	var members []model.GroupMember
	if err := database.DB.Where("entity_id = ?", entityID).Find(&members).Error; err != nil {
		return []model.Group{}, err
	}
	if len(members) == 0 {
		return []model.Group{}, nil
	}
	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.GroupID)
	}
	groups := []model.Group{}
	if err := database.DB.Where("id IN ?", ids).Find(&groups).Error; err != nil {
		return []model.Group{}, err
	}
	return groups, nil
}
