package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
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

func GetJoinRequestsByGroup(groupID string) ([]model.GroupJoinRequest, error) {
	var requests []model.GroupJoinRequest
	if err := database.DB.Where("group_id = ?", groupID).Find(&requests).Error; err != nil {
		return []model.GroupJoinRequest{}, err
	}
	for i := range requests {
		PopulateJoinRequest(&requests[i])
	}
	return requests, nil
}

func GetJoinRequestsByEntity(entityID string) ([]model.GroupJoinRequest, error) {
	var requests []model.GroupJoinRequest
	if err := database.DB.Where("entity_id = ?", entityID).Find(&requests).Error; err != nil {
		return []model.GroupJoinRequest{}, err
	}
	for i := range requests {
		PopulateJoinRequest(&requests[i])
	}
	return requests, nil
}

func GetJoinRequestByID(id string) (model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	if err := database.DB.Where("id = ?", id).First(&request).Error; err != nil {
		return model.GroupJoinRequest{}, err
	}
	PopulateJoinRequest(&request)
	return request, nil
}

func CreateJoinRequest(request model.GroupJoinRequest) (model.GroupJoinRequest, error) {
	if request.ID == "" {
		request.ID = ulid.Make().Prefixed("gjr")
	}
	if err := database.DB.Create(&request).Error; err != nil {
		return model.GroupJoinRequest{}, err
	}
	PopulateJoinRequest(&request)
	return request, nil
}

func UpdateJoinRequest(request model.GroupJoinRequest) (model.GroupJoinRequest, error) {
	if err := database.DB.Save(&request).Error; err != nil {
		return model.GroupJoinRequest{}, err
	}
	PopulateJoinRequest(&request)
	return request, nil
}

func DeleteJoinRequest(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.GroupJoinRequest{}).Error; err != nil {
		return err
	}
	return nil
}

func PopulateJoinRequest(request *model.GroupJoinRequest) {
	comments, err := GetCommentsForJoinRequest(request.ID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get comments for join request %s: %v", request.ID, err)
	}
	request.Comments = comments
}

func GetCommentsForJoinRequest(requestID string) ([]model.GroupJoinRequestComment, error) {
	var comments []model.GroupJoinRequestComment
	if err := database.DB.Where("request_id = ?", requestID).Find(&comments).Error; err != nil {
		return []model.GroupJoinRequestComment{}, err
	}
	return comments, nil
}

func CreateJoinRequestComment(comment model.GroupJoinRequestComment) (model.GroupJoinRequestComment, error) {
	if comment.ID == "" {
		comment.ID = ulid.Make().Prefixed("gjrc")
	}
	if err := database.DB.Create(&comment).Error; err != nil {
		return model.GroupJoinRequestComment{}, err
	}
	return comment, nil
}

func DeleteJoinRequestComment(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.GroupJoinRequestComment{}).Error; err != nil {
		return err
	}
	return nil
}
