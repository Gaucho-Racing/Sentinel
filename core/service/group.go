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

func GetMembersForGroup(groupID string) ([]model.GroupMember, error) {
	var members []model.GroupMember
	if err := database.DB.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return []model.GroupMember{}, err
	}
	return members, nil
}

func GetGroupMember(groupID string, entityID string) (model.GroupMember, error) {
	var member model.GroupMember
	if err := database.DB.Where("group_id = ? AND entity_id = ?", groupID, entityID).First(&member).Error; err != nil {
		return model.GroupMember{}, err
	}
	return member, nil
}

func CreateGroupMember(member model.GroupMember) (model.GroupMember, error) {
	if err := database.DB.Create(&member).Error; err != nil {
		return model.GroupMember{}, err
	}
	return member, nil
}

func UpdateGroupMember(member model.GroupMember) (model.GroupMember, error) {
	if err := database.DB.Save(&member).Error; err != nil {
		return model.GroupMember{}, err
	}
	return member, nil
}

func DeleteGroupMember(groupID string, entityID string) error {
	if err := database.DB.Where("group_id = ? AND entity_id = ?", groupID, entityID).Delete(&model.GroupMember{}).Error; err != nil {
		return err
	}
	return nil
}

func GetOwnersForGroup(groupID string) ([]model.GroupOwner, error) {
	var owners []model.GroupOwner
	if err := database.DB.Where("group_id = ?", groupID).Find(&owners).Error; err != nil {
		return []model.GroupOwner{}, err
	}
	return owners, nil
}

func GetGroupOwner(groupID string, entityID string) (model.GroupOwner, error) {
	var owner model.GroupOwner
	if err := database.DB.Where("group_id = ? AND entity_id = ?", groupID, entityID).First(&owner).Error; err != nil {
		return model.GroupOwner{}, err
	}
	return owner, nil
}

func CreateGroupOwner(owner model.GroupOwner) (model.GroupOwner, error) {
	if err := database.DB.Create(&owner).Error; err != nil {
		return model.GroupOwner{}, err
	}
	return owner, nil
}

func DeleteGroupOwner(groupID string, entityID string) error {
	if err := database.DB.Where("group_id = ? AND entity_id = ?", groupID, entityID).Delete(&model.GroupOwner{}).Error; err != nil {
		return err
	}
	return nil
}
