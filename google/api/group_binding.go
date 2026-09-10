package api

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/sentinel/google/service"
	"github.com/gin-gonic/gin"
)

// ListGoogleBindings returns all group→Google-Group bindings, optionally
// filtered to a single group_id. Used by the web UI and by reconciliation.
func ListGoogleBindings(c *gin.Context) {
	Require(c, RequestTokenHasInternalAccess(c) || RequestTokenHasFirstPartyAccess(c))

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
	GroupID                               string `json:"group_id" binding:"required"`
	GoogleGroupEmail                      string `json:"google_group_email" binding:"required"`
	ConfirmOverwriteRequestedGroupMembers bool   `json:"confirm_overwrite_requested_group_members"`
}

func CreateGoogleBinding(c *gin.Context) {
	var req createGoogleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenHasAdminAccess(c))
	email, err := normalizeGoogleGroupEmail(req.GoogleGroupEmail, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google_group_email must be a valid email address"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	binding, preflight, err := service.QueueGoogleBinding(
		ctx,
		req.GroupID,
		email,
		"",
		service.GoogleBindingConfirmations{
			OverwriteRequestedGroupMembers: req.ConfirmOverwriteRequestedGroupMembers,
		},
	)
	if err != nil {
		writeGoogleBindingError(c, err, preflight)
		return
	}
	writeGoogleBindingQueued(c, binding)
}

type googleBindingPreflightRequest struct {
	GroupID          string `json:"group_id" binding:"required"`
	GoogleGroupEmail string `json:"google_group_email"`
}

func PreflightGoogleBinding(c *gin.Context) {
	var req googleBindingPreflightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenHasAdminAccess(c))
	email, err := normalizeGoogleGroupEmail(req.GoogleGroupEmail, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google_group_email must be a valid email address"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	preflight, err := service.PreflightGoogleBinding(ctx, req.GroupID, email)
	if err != nil {
		writeGoogleBindingError(c, err, preflight)
		return
	}
	c.JSON(http.StatusOK, preflight)
}

type applyGoogleBindingRequest struct {
	GroupID                               string `json:"group_id" binding:"required"`
	GoogleGroupEmail                      string `json:"google_group_email"`
	ExpectedCurrentBindingID              string `json:"expected_current_binding_id"`
	ConfirmOverwriteRequestedGroupMembers bool   `json:"confirm_overwrite_requested_group_members"`
	ConfirmDeletePreviousGroup            bool   `json:"confirm_delete_previous_group"`
}

func ApplyGoogleBinding(c *gin.Context) {
	var req applyGoogleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenHasAdminAccess(c))
	email, err := normalizeGoogleGroupEmail(req.GoogleGroupEmail, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google_group_email must be a valid email address"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	binding, preflight, err := service.QueueGoogleBinding(
		ctx,
		req.GroupID,
		email,
		req.ExpectedCurrentBindingID,
		service.GoogleBindingConfirmations{
			OverwriteRequestedGroupMembers: req.ConfirmOverwriteRequestedGroupMembers,
			DeletePreviousGroup:            req.ConfirmDeletePreviousGroup,
		},
	)
	if err != nil {
		writeGoogleBindingError(c, err, preflight)
		return
	}
	writeGoogleBindingQueued(c, binding)
}

func normalizeGoogleGroupEmail(value string, allowEmpty bool) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" && allowEmpty {
		return "", nil
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", errors.New("invalid email address")
	}
	return email, nil
}

func writeGoogleBindingError(c *gin.Context, err error, preflight service.GoogleBindingPreflight) {
	var confirmationErr *service.ConfirmationRequiredError
	var stateChangedErr *service.BindingStateChangedError
	var alreadyBoundErr *service.GoogleGroupAlreadyBoundError
	switch {
	case errors.As(err, &confirmationErr), errors.As(err, &stateChangedErr):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "preflight": preflight})
	case errors.As(err, &alreadyBoundErr):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrBindingSyncInProgress):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrGoogleSyncUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func writeGoogleBindingQueued(c *gin.Context, binding *model.GroupGoogleBinding) {
	if binding == nil || binding.OperationID == "" {
		c.JSON(http.StatusOK, gin.H{"binding": binding, "status": "unchanged"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"binding":      binding,
		"message":      "google group sync queued",
		"operation_id": binding.OperationID,
		"status":       "queued",
	})
}

// DeleteGoogleBinding removes a binding by ID. The group_id query param is
// required to scope the delete — protects against URL tampering that would
// otherwise let a caller delete a binding for a group they don't control.
func DeleteGoogleBinding(c *gin.Context) {
	bindingID := c.Param("bindingID")
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id query param is required"})
		return
	}
	Require(c, RequestTokenHasAdminAccess(c))
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	binding, preflight, err := service.QueueGoogleBinding(
		ctx,
		groupID,
		"",
		bindingID,
		service.GoogleBindingConfirmations{
			DeletePreviousGroup: c.Query("confirm_delete_group") == "true",
		},
	)
	if err != nil {
		writeGoogleBindingError(c, err, preflight)
		return
	}
	writeGoogleBindingQueued(c, binding)
}
