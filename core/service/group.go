package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

// AdminsGroupID is the fixed ID of the global Admins group. Members get
// owner-equivalent permissions on every group and other admin-gated surfaces.
const AdminsGroupID = "grp_01kqs3w6h82xkdnft94vpj7qrm"

// IsAdmin reports whether the given entity is a member of the Admins group.
// Returns false if the lookup fails so callers can treat it as a deny-by-default.
func IsAdmin(entityID string) bool {
	if entityID == "" {
		return false
	}
	_, err := GetGroupMember(AdminsGroupID, entityID)
	return err == nil
}

func GetAllGroups() ([]model.Group, error) {
	groups := []model.Group{}
	if err := database.DB.Find(&groups).Error; err != nil {
		return []model.Group{}, err
	}
	for i := range groups {
		PopulateGroup(&groups[i])
	}
	return groups, nil
}

func GetGroupByID(id string) (model.Group, error) {
	var group model.Group
	if err := database.DB.Where("id = ?", id).First(&group).Error; err != nil {
		return model.Group{}, err
	}
	PopulateGroup(&group)
	return group, nil
}

func PopulateGroup(group *model.Group) {
	if err := database.DB.Model(&model.GroupMember{}).Where("group_id = ?", group.ID).Count(&group.MemberCount).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to count members for group %s: %v", group.ID, err)
	}
	if err := database.DB.Model(&model.GroupOwner{}).Where("group_id = ?", group.ID).Count(&group.OwnerCount).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to count owners for group %s: %v", group.ID, err)
	}
	if err := database.DB.Model(&model.GroupJoinRequest{}).Where("group_id = ? AND status = ?", group.ID, model.GroupJoinRequestStatusPending).Count(&group.PendingCount).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to count pending requests for group %s: %v", group.ID, err)
	}
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

// IsGroupNameAvailable reports whether the given name can be used by a new or
// updated group. Case-insensitive. Pass excludeID to allow a group to keep its
// own current name during an update.
func IsGroupNameAvailable(name string, excludeID string) (bool, error) {
	var count int64
	q := database.DB.Model(&model.Group{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

func DeleteGroup(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.Group{}).Error; err != nil {
		return err
	}
	return nil
}

func GetMembersForGroup(groupID string) ([]model.GroupMember, error) {
	members := []model.GroupMember{}
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
	owners := []model.GroupOwner{}
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
	requests := []model.GroupJoinRequest{}
	if err := database.DB.Where("group_id = ?", groupID).Find(&requests).Error; err != nil {
		return []model.GroupJoinRequest{}, err
	}
	for i := range requests {
		PopulateJoinRequest(&requests[i])
	}
	return requests, nil
}

func GetJoinRequestsByEntity(entityID string) ([]model.GroupJoinRequest, error) {
	requests := []model.GroupJoinRequest{}
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
	comments := []model.GroupJoinRequestComment{}
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

// Discord role bindings

func GetDiscordBindingsForGroup(groupID string) ([]model.GroupDiscordRoleBinding, error) {
	bindings := []model.GroupDiscordRoleBinding{}
	if err := database.DB.Where("group_id = ?", groupID).Find(&bindings).Error; err != nil {
		return []model.GroupDiscordRoleBinding{}, err
	}
	return bindings, nil
}

func CreateGroupDiscordBinding(binding model.GroupDiscordRoleBinding) (model.GroupDiscordRoleBinding, error) {
	if binding.ID == "" {
		binding.ID = ulid.Make().Prefixed("gdrb")
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		return model.GroupDiscordRoleBinding{}, err
	}
	return binding, nil
}

// DeleteGroupDiscordBinding scopes the delete to (groupID, bindingID) so a
// URL-tampered request can't drop a binding from a different group.
func DeleteGroupDiscordBinding(groupID string, bindingID string) error {
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, bindingID).Delete(&model.GroupDiscordRoleBinding{}).Error; err != nil {
		return err
	}
	return nil
}
