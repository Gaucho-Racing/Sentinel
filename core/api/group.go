package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// validateMembershipExpiration enforces the 1-year cap on time-boxed
// memberships and join requests. Uses AddDate(1, 0, 0) so leap years are
// handled correctly, with a small slack for clock drift between client and
// server.
func validateMembershipExpiration(hasExpiration bool, expiresAt time.Time) error {
	if !hasExpiration {
		return nil
	}
	if expiresAt.IsZero() {
		return errors.New("expires_at is required when has_expiration is true")
	}
	now := time.Now()
	if !expiresAt.After(now) {
		return errors.New("expires_at must be in the future")
	}
	maxAllowed := now.AddDate(1, 0, 0).Add(1 * time.Minute)
	if expiresAt.After(maxAllowed) {
		return errors.New("maximum membership duration is 1 year")
	}
	return nil
}

func GetAllGroups(c *gin.Context) {
	groups, err := service.GetAllGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func GetGroupByID(c *gin.Context) {
	id := c.Param("id")
	group, err := service.GetGroupByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

type upsertGroupRequest struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	AllowedSources []string `json:"allowed_sources"`
}

func CreateOrUpdateGroup(c *gin.Context) {
	var req upsertGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := service.GetGroupByID(req.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	available, err := service.IsGroupNameAvailable(req.Name, existing.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !available {
		c.JSON(http.StatusConflict, gin.H{"error": "a group with that name already exists"})
		return
	}

	var group model.Group
	if existing.ID != "" {
		// Preserve CreatedBy and CreatedAt so updates don't overwrite the
		// original creator lineage.
		group = existing
		group.Name = req.Name
		group.Description = req.Description
		group.AllowedSources = model.StringSlice(req.AllowedSources)
		group, err = service.UpdateGroup(group)
	} else {
		group = model.Group{
			ID:             req.ID,
			Name:           req.Name,
			Description:    req.Description,
			AllowedSources: model.StringSlice(req.AllowedSources),
			CreatedBy:      GetRequestTokenEntityID(c),
		}
		group, err = service.CreateGroup(group)
		// Auto-add the creator as an owner so new groups aren't ownerless.
		// Skip when there's no auth context — fabricating a row with an
		// empty entity_id would just create a dangling owner.
		if err == nil && group.CreatedBy != "" {
			if _, ownerErr := service.CreateGroupOwner(model.GroupOwner{
				GroupID:  group.ID,
				EntityID: group.CreatedBy,
				AddedBy:  group.CreatedBy,
			}); ownerErr != nil {
				logger.SugarLogger.Errorf("Failed to auto-add creator as owner of group %s: %v", group.ID, ownerErr)
			}
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func GetGroupApplications(c *gin.Context) {
	id := c.Param("id")
	apps, err := service.GetApplicationsForGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteGroup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

// Members

func GetGroupMembers(c *gin.Context) {
	id := c.Param("id")
	members, err := service.GetMembersForGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

type addGroupMemberRequest struct {
	EntityID      string    `json:"entity_id" binding:"required"`
	Source        string    `json:"source"`
	AddedBy       string    `json:"added_by"`
	HasExpiration bool      `json:"has_expiration"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func AddGroupMember(c *gin.Context) {
	id := c.Param("id")
	var req addGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateMembershipExpiration(req.HasExpiration, req.ExpiresAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	member, err := service.CreateGroupMember(model.GroupMember{
		GroupID:       id,
		EntityID:      req.EntityID,
		Source:        req.Source,
		AddedBy:       req.AddedBy,
		HasExpiration: req.HasExpiration,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

func RemoveGroupMember(c *gin.Context) {
	id := c.Param("id")
	entityID := c.Param("entityID")
	if err := service.DeleteGroupMember(id, entityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed from group"})
}

// Owners

func GetGroupOwners(c *gin.Context) {
	id := c.Param("id")
	owners, err := service.GetOwnersForGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, owners)
}

type addGroupOwnerRequest struct {
	EntityID string `json:"entity_id" binding:"required"`
	AddedBy  string `json:"added_by"`
}

func AddGroupOwner(c *gin.Context) {
	id := c.Param("id")
	var req addGroupOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	owner, err := service.CreateGroupOwner(model.GroupOwner{
		GroupID:  id,
		EntityID: req.EntityID,
		AddedBy:  req.AddedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, owner)
}

func RemoveGroupOwner(c *gin.Context) {
	id := c.Param("id")
	entityID := c.Param("entityID")
	if err := service.DeleteGroupOwner(id, entityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "owner removed from group"})
}

// Join Requests

func GetGroupJoinRequests(c *gin.Context) {
	id := c.Param("id")
	requests, err := service.GetJoinRequestsByGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func GetGroupJoinRequest(c *gin.Context) {
	requestID := c.Param("requestID")
	request, err := service.GetJoinRequestByID(requestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "join request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

type createJoinRequestRequest struct {
	EntityID      string    `json:"entity_id" binding:"required"`
	HasExpiration bool      `json:"has_expiration"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func CreateGroupJoinRequest(c *gin.Context) {
	id := c.Param("id")
	var req createJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateMembershipExpiration(req.HasExpiration, req.ExpiresAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := service.GetGroupMember(id, req.EntityID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "entity is already a member of this group"})
		return
	}
	request, err := service.CreateJoinRequest(model.GroupJoinRequest{
		GroupID:       id,
		EntityID:      req.EntityID,
		Status:        string(model.GroupJoinRequestStatusPending),
		HasExpiration: req.HasExpiration,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

type reviewJoinRequestRequest struct {
	ReviewedBy string `json:"reviewed_by" binding:"required"`
	// Optional approval-time overrides. When provided, they replace the
	// expiration that the requester originally chose — used by reviewers
	// who want to grant a shorter/longer membership than what was asked
	// for. Both fields propagate to the created GroupMember and the
	// request row is updated for audit lineage.
	HasExpiration *bool      `json:"has_expiration,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

func ApproveGroupJoinRequest(c *gin.Context) {
	requestID := c.Param("requestID")
	var req reviewJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request, err := service.GetJoinRequestByID(requestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "join request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasExpiration := request.HasExpiration
	expiresAt := request.ExpiresAt
	if req.HasExpiration != nil {
		hasExpiration = *req.HasExpiration
	}
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}
	if err := validateMembershipExpiration(hasExpiration, expiresAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.Status = string(model.GroupJoinRequestStatusApproved)
	request.ReviewedBy = req.ReviewedBy
	request.ReviewedAt = time.Now()
	request.HasExpiration = hasExpiration
	request.ExpiresAt = expiresAt
	request, err = service.UpdateJoinRequest(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = service.CreateGroupMember(model.GroupMember{
		GroupID:       request.GroupID,
		EntityID:      request.EntityID,
		Source:        string(model.GroupMemberSourceDirect),
		AddedBy:       req.ReviewedBy,
		HasExpiration: hasExpiration,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

func RejectGroupJoinRequest(c *gin.Context) {
	requestID := c.Param("requestID")
	var req reviewJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request, err := service.GetJoinRequestByID(requestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "join request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	request.Status = string(model.GroupJoinRequestStatusRejected)
	request.ReviewedBy = req.ReviewedBy
	request.ReviewedAt = time.Now()
	request, err = service.UpdateJoinRequest(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

func DeleteGroupJoinRequest(c *gin.Context) {
	requestID := c.Param("requestID")
	if err := service.DeleteJoinRequest(requestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "join request deleted"})
}

// Join Request Comments

type createJoinRequestCommentRequest struct {
	EntityID string `json:"entity_id" binding:"required"`
	Comment  string `json:"comment" binding:"required"`
}

func CreateJoinRequestComment(c *gin.Context) {
	requestID := c.Param("requestID")
	var req createJoinRequestCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment, err := service.CreateJoinRequestComment(model.GroupJoinRequestComment{
		RequestID: requestID,
		EntityID:  req.EntityID,
		Comment:   req.Comment,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
}

func DeleteJoinRequestComment(c *gin.Context) {
	commentID := c.Param("commentID")
	if err := service.DeleteJoinRequestComment(commentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

// Discord role bindings

func GetGroupDiscordBindings(c *gin.Context) {
	id := c.Param("id")
	bindings, err := service.GetDiscordBindingsForGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bindings)
}

type addDiscordBindingRequest struct {
	DiscordRoleIDs []string `json:"discord_role_ids" binding:"required"`
}

func AddGroupDiscordBinding(c *gin.Context) {
	id := c.Param("id")
	var req addDiscordBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.DiscordRoleIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discord_role_ids must be non-empty"})
		return
	}
	binding, err := service.CreateGroupDiscordBinding(model.GroupDiscordRoleBinding{
		GroupID:        id,
		DiscordRoleIDs: model.StringSlice(req.DiscordRoleIDs),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, binding)
}

func RemoveGroupDiscordBinding(c *gin.Context) {
	id := c.Param("id")
	bindingID := c.Param("bindingID")
	if err := service.DeleteGroupDiscordBinding(id, bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "discord binding removed"})
}
