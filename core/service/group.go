package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllGroups() ([]model.Group, error) {
	var groups []model.Group
	if err := database.DB.Find(&groups).Error; err != nil {
		return []model.Group{}, err
	}
	return groups, nil
}

func GetGroupByID(id string) (model.Group, error) {
	var group model.Group
	if err := database.DB.Where("id = ?", id).First(&group).Error; err != nil {
		return model.Group{}, err
	}
	return group, nil
}

func CreateGroup(group model.Group) (model.Group, error) {
	if group.ID == "" {
		group.ID = ulid.Make().Prefixed("grp")
	}
	if err := database.DB.Create(&group).Error; err != nil {
		return model.Group{}, err
	}
	return group, nil
}

func UpdateGroup(group model.Group) (model.Group, error) {
	if err := database.DB.Save(&group).Error; err != nil {
		return model.Group{}, err
	}
	return group, nil
}

func DeleteGroup(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.Group{}).Error; err != nil {
		return err
	}
	return nil
}
