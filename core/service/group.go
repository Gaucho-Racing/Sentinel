package service

import (
	"errors"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdminsGroupID is the fixed ID of the global Admins group. Members get
// owner-equivalent permissions on every group and other admin-gated surfaces.
const AdminsGroupID = "grp_01kqs3w6h82xkdnft94vpj7qrm"

var (
	ErrJoinRequestNotPending = errors.New("join request is not pending")
	ErrGroupMemberExists     = errors.New("entity is already a member of this group")
)

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

// GetMembershipsForEntity returns every GroupMember row for an entity,
// optionally filtered to a single Source. An empty source returns all rows.
func GetMembershipsForEntity(entityID, source string) ([]model.GroupMember, error) {
	members := []model.GroupMember{}
	q := database.DB.Where("entity_id = ?", entityID)
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if err := q.Find(&members).Error; err != nil {
		return []model.GroupMember{}, err
	}
	return members, nil
}

// DeleteGroupMember removes the membership row for (groupID, entityID).
// If source is non-empty, the delete is scoped to that source — preventing
// callers from accidentally removing a row written by a different source
// (e.g. discord reconciliation deleting a DIRECT membership).
func DeleteGroupMember(groupID, entityID, source string) error {
	q := database.DB.Where("group_id = ? AND entity_id = ?", groupID, entityID)
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if err := q.Delete(&model.GroupMember{}).Error; err != nil {
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

func GetJoinRequestForGroup(groupID string, id string) (model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	if err := database.DB.Where("id = ? AND group_id = ?", id, groupID).First(&request).Error; err != nil {
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

func ReviewJoinRequest(groupID string, id string, reviewerID string, status model.GroupJoinRequestStatus, hasExpiration bool, expiresAt time.Time) (model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND group_id = ?", id, groupID).
			First(&request).Error; err != nil {
			return err
		}
		if request.Status != string(model.GroupJoinRequestStatusPending) {
			return ErrJoinRequestNotPending
		}

		if status == model.GroupJoinRequestStatusApproved {
			member := model.GroupMember{
				GroupID:       request.GroupID,
				EntityID:      request.EntityID,
				Source:        string(model.GroupMemberSourceDirect),
				AddedBy:       reviewerID,
				HasExpiration: hasExpiration,
				ExpiresAt:     expiresAt,
			}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrGroupMemberExists
			}
		}

		request.Status = string(status)
		request.ReviewedBy = reviewerID
		request.ReviewedAt = time.Now()
		request.HasExpiration = hasExpiration
		request.ExpiresAt = expiresAt
		return tx.Save(&request).Error
	})
	if err != nil {
		return model.GroupJoinRequest{}, err
	}
	PopulateJoinRequest(&request)
	return request, nil
}

func DeleteJoinRequestForGroup(groupID string, id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var request model.GroupJoinRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND group_id = ?", id, groupID).
			First(&request).Error; err != nil {
			return err
		}
		if err := tx.Where("request_id = ?", request.ID).Delete(&model.GroupJoinRequestComment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&request).Error
	})
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

// GetJoinRequestComment returns a single comment by ID. Used by the
// delete handler to authorize the requester against the comment's
// claimed author before letting them delete it.
func GetJoinRequestCommentForRequest(requestID string, id string) (model.GroupJoinRequestComment, error) {
	var comment model.GroupJoinRequestComment
	if err := database.DB.Where("id = ? AND request_id = ?", id, requestID).First(&comment).Error; err != nil {
		return model.GroupJoinRequestComment{}, err
	}
	return comment, nil
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

func DeleteJoinRequestCommentForRequest(requestID string, id string) error {
	result := database.DB.Where("id = ? AND request_id = ?", id, requestID).Delete(&model.GroupJoinRequestComment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteMembersBySource removes all GroupMember rows in the given group that
// were added via the given source. Used when a group revokes a source (e.g.
// DISCORD is unchecked) — anyone who was only there because of that source
// loses access; DIRECT members are untouched.
func DeleteMembersBySource(groupID string, source string) error {
	if err := database.DB.Where("group_id = ? AND source = ?", groupID, source).Delete(&model.GroupMember{}).Error; err != nil {
		return err
	}
	return nil
}
