package api

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/sentinel/google/service"
	"github.com/gin-gonic/gin"
)

// ListGoogleBindings returns all group→Google-Group bindings, optionally
// filtered to a single group_id. Used by the web UI and by reconciliation.
func ListGoogleBindings(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	if groupID := c.Query("group_id"); groupID != "" {
		binding, err := service.GetGoogleBindingForGroup(groupID)
		if errors.Is(err, service.ErrBindingNotFound) {
			c.JSON(http.StatusOK, []model.GroupGoogleBinding{})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, []model.GroupGoogleBinding{binding})
		return
	}

	bindings, err := service.GetAllGoogleBindings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bindings)
}

type createGoogleBindingRequest struct {
	GroupID          string `json:"group_id" binding:"required"`
	GoogleGroupEmail string `json:"google_group_email" binding:"required"`
}

func CreateGoogleBinding(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	var req createGoogleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.TrimSpace(req.GoogleGroupEmail)
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google_group_email must be a valid email address"})
		return
	}

	binding, err := service.CreateGoogleBinding(model.GroupGoogleBinding{
		GroupID:          req.GroupID,
		GoogleGroupEmail: email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, binding)
}

// DeleteGoogleBinding removes a binding by ID. The group_id query param is
// required to scope the delete — protects against URL tampering that would
// otherwise let a caller delete a binding for a group they don't control.
func DeleteGoogleBinding(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))

	bindingID := c.Param("bindingID")
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id query param is required"})
		return
	}
	if err := service.DeleteGoogleBinding(groupID, bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
