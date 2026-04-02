package api

import (
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func CreateOrUpdateGroup(c *gin.Context) {
	var group model.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, err := service.GetGroupByID(group.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.ID != "" {
		group, err = service.UpdateGroup(group)
	} else {
		group, err = service.CreateGroup(group)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
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
	request, err := service.CreateJoinRequest(model.GroupJoinRequest{
		GroupID:       id,
		EntityID:      req.EntityID,
		Status:        "PENDING",
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
	request.Status = "APPROVED"
	request.ReviewedBy = req.ReviewedBy
	request.ReviewedAt = time.Now()
	request, err = service.UpdateJoinRequest(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = service.CreateGroupMember(model.GroupMember{
		GroupID:       request.GroupID,
		EntityID:      request.EntityID,
		Source:        "JOIN_REQUEST",
		AddedBy:       req.ReviewedBy,
		HasExpiration: request.HasExpiration,
		ExpiresAt:     request.ExpiresAt,
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
	request.Status = "REJECTED"
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
