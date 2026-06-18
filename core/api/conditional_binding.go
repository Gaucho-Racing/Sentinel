package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

func GetGroupConditionalBindings(c *gin.Context) {
	Require(c, RequestTokenExists(c))

	id := c.Param("id")
	bindings, err := service.GetConditionalBindingsForGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bindings)
}

type createConditionalBindingRequest struct {
	RequiredGroupIDs []string `json:"required_group_ids" binding:"required"`
}

func CreateGroupConditionalBinding(c *gin.Context) {
	id := c.Param("id")
	if !requireGroupOwnerOrAdmin(c, id) {
		return
	}
	var req createConditionalBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.RequiredGroupIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "required_group_ids must not be empty"})
		return
	}

	binding, err := service.CreateConditionalBinding(model.GroupConditionalBinding{
		GroupID:          id,
		RequiredGroupIDs: model.StringSlice(req.RequiredGroupIDs),
	})
	if err != nil {
		// Cycle / self-ref are validation failures, not server errors — 400
		// so admins see a clear "won't work" rather than a 500.
		if errors.Is(err, service.ErrConditionalBindingCycle) || errors.Is(err, service.ErrConditionalBindingSelfRef) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// New binding may newly-satisfy entities we haven't seen yet — kick a
	// full sweep so they get the membership without waiting for the cron.
	service.TriggerReconcileAllConditional()
	c.JSON(http.StatusOK, binding)
}

func DeleteGroupConditionalBinding(c *gin.Context) {
	id := c.Param("id")
	if !requireGroupOwnerOrAdmin(c, id) {
		return
	}
	bindingID := c.Param("bindingID")
	if err := service.DeleteConditionalBinding(id, bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Removing a binding may newly-DISqualify entities — sweep to strip
	// their now-orphaned CONDITIONAL memberships.
	service.TriggerReconcileAllConditional()
	c.JSON(http.StatusOK, gin.H{"message": "conditional binding deleted"})
}
