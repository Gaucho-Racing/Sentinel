package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/service"
	"github.com/gin-gonic/gin"
)

// ListRoleBindings returns all role bindings, optionally filtered by group_id.
// Used by the web UI (per-group view) and by reconciliation (full sweep).
func ListRoleBindings(c *gin.Context) {
	groupID := c.Query("group_id")
	var (
		bindings []model.GroupDiscordRoleBinding
		err      error
	)
	if groupID != "" {
		bindings, err = service.GetRoleBindingsForGroup(groupID)
	} else {
		bindings, err = service.GetAllRoleBindings()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bindings)
}

type createRoleBindingRequest struct {
	GroupID        string   `json:"group_id" binding:"required"`
	DiscordRoleIDs []string `json:"discord_role_ids" binding:"required"`
}

func CreateRoleBinding(c *gin.Context) {
	var req createRoleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.DiscordRoleIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discord_role_ids must be non-empty"})
		return
	}
	binding, err := service.CreateRoleBinding(model.GroupDiscordRoleBinding{
		GroupID:        req.GroupID,
		DiscordRoleIDs: model.StringSlice(req.DiscordRoleIDs),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	service.TriggerReconcileAll()
	c.JSON(http.StatusOK, binding)
}

// DeleteRoleBinding removes a binding by ID. The group_id query param is
// required to scope the delete — protects against URL tampering that would
// otherwise let a caller delete a binding they don't own access to.
func DeleteRoleBinding(c *gin.Context) {
	bindingID := c.Param("bindingID")
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id query param is required"})
		return
	}
	if err := service.DeleteRoleBinding(groupID, bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	service.TriggerReconcileAll()
	c.Status(http.StatusNoContent)
}
