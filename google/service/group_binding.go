package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gaucho-racing/sentinel/google/database"
	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
	"github.com/gaucho-racing/sentinel/google/pkg/sentinel"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

var ErrBindingNotFound = errors.New("group google binding not found")
var ErrGoogleSyncUnavailable = errors.New("google group management is not configured")
var ErrBindingSyncInProgress = errors.New("google group binding sync is already in progress")
var ErrGoogleGroupStateChanged = errors.New("google group changed after preflight")

const ManagedGoogleGroupOwnerEmail = "team@gauchoracing.com"
const BindingStatusActive = "active"
const BindingStatusPending = "pending"

type GoogleGroupMemberSnapshot struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type GoogleGroupSnapshot struct {
	RequestedEmail string                      `json:"requested_email"`
	ID             string                      `json:"id,omitempty"`
	Email          string                      `json:"email"`
	Name           string                      `json:"name,omitempty"`
	Exists         bool                        `json:"exists"`
	Members        []GoogleGroupMemberSnapshot `json:"members"`
}

type GoogleBindingConfirmations struct {
	OverwriteRequestedGroupMembers bool `json:"overwrite_requested_group_members"`
	DeletePreviousGroup            bool `json:"delete_previous_group"`
}

type GoogleBindingPreflight struct {
	GroupID              string                     `json:"group_id"`
	RequestedEmail       string                     `json:"requested_email"`
	BindingChanged       bool                       `json:"binding_changed"`
	CurrentBinding       *model.GroupGoogleBinding  `json:"current_binding"`
	PreviousGroup        *GoogleGroupSnapshot       `json:"previous_group"`
	RequestedGroup       *GoogleGroupSnapshot       `json:"requested_group"`
	RequiredConfirmation GoogleBindingConfirmations `json:"required_confirmation"`
}

type ConfirmationRequiredError struct {
	Preflight GoogleBindingPreflight
}

func (e *ConfirmationRequiredError) Error() string {
	return "google group binding confirmation is required"
}

type BindingStateChangedError struct {
	Preflight GoogleBindingPreflight
}

func (e *BindingStateChangedError) Error() string {
	return "google group binding changed after preflight"
}

type GoogleGroupAlreadyBoundError struct {
	Binding model.GroupGoogleBinding
}

func (e *GoogleGroupAlreadyBoundError) Error() string {
	return fmt.Sprintf("google group %s is already bound to sentinel group %s", e.Binding.GoogleGroupEmail, e.Binding.GroupID)
}

var bindingMutationMu sync.Mutex

func GetAllGoogleBindings() ([]model.GroupGoogleBinding, error) {
	bindings := []model.GroupGoogleBinding{}
	if err := database.DB.Find(&bindings).Error; err != nil {
		return []model.GroupGoogleBinding{}, err
	}
	return bindings, nil
}

// GetGoogleBindingForGroup returns the binding for a group, or
// ErrBindingNotFound if the group has none. The 1:1 model means at most one
// row per group.
func GetGoogleBindingForGroup(groupID string) (model.GroupGoogleBinding, error) {
	var binding model.GroupGoogleBinding
	if err := database.DB.Where("group_id = ?", groupID).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GroupGoogleBinding{}, ErrBindingNotFound
		}
		return model.GroupGoogleBinding{}, err
	}
	return binding, nil
}

func getGoogleBindingForEmail(googleGroupEmail string) (model.GroupGoogleBinding, error) {
	var binding model.GroupGoogleBinding
	if err := database.DB.Where("LOWER(google_group_email) = ?", strings.ToLower(googleGroupEmail)).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GroupGoogleBinding{}, ErrBindingNotFound
		}
		return model.GroupGoogleBinding{}, err
	}
	return binding, nil
}

func inspectGoogleGroup(ctx context.Context, groupEmail string) (*GoogleGroupSnapshot, error) {
	group, exists, err := getGoogleGroup(ctx, groupEmail)
	if err != nil {
		return nil, err
	}
	snapshot := &GoogleGroupSnapshot{
		RequestedEmail: groupEmail,
		Email:          groupEmail,
		Exists:         exists,
		Members:        []GoogleGroupMemberSnapshot{},
	}
	if !exists {
		return snapshot, nil
	}
	snapshot.ID = group.Id
	snapshot.Email = strings.ToLower(group.Email)
	snapshot.Name = group.Name
	members, err := listGroupMembers(ctx, groupEmail)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		snapshot.Members = append(snapshot.Members, GoogleGroupMemberSnapshot{
			Email: strings.ToLower(member.Email),
			Role:  member.Role,
		})
	}
	return snapshot, nil
}

func PreflightGoogleBinding(ctx context.Context, groupID, requestedEmail string) (GoogleBindingPreflight, error) {
	if directorySvc == nil {
		return GoogleBindingPreflight{}, ErrGoogleSyncUnavailable
	}
	requestedEmail = strings.ToLower(strings.TrimSpace(requestedEmail))
	preflight := GoogleBindingPreflight{
		GroupID:        groupID,
		RequestedEmail: requestedEmail,
	}
	current, err := GetGoogleBindingForGroup(groupID)
	if err != nil && !errors.Is(err, ErrBindingNotFound) {
		return GoogleBindingPreflight{}, err
	}
	if err == nil {
		preflight.CurrentBinding = &current
	}
	currentEmail := ""
	if preflight.CurrentBinding != nil {
		currentEmail = strings.ToLower(preflight.CurrentBinding.GoogleGroupEmail)
		preflight.PreviousGroup, err = inspectGoogleGroup(ctx, currentEmail)
		if err != nil {
			return GoogleBindingPreflight{}, err
		}
	}
	preflight.BindingChanged = currentEmail != requestedEmail
	if requestedEmail != "" {
		preflight.RequestedGroup, err = inspectGoogleGroup(ctx, requestedEmail)
		if err != nil {
			return GoogleBindingPreflight{}, err
		}
		bindingEmail := requestedEmail
		if preflight.RequestedGroup.Exists {
			bindingEmail = preflight.RequestedGroup.Email
		}
		bound, bindingErr := getGoogleBindingForEmail(bindingEmail)
		if bindingErr != nil && !errors.Is(bindingErr, ErrBindingNotFound) {
			return GoogleBindingPreflight{}, bindingErr
		}
		if bindingErr == nil && bound.GroupID != groupID {
			return GoogleBindingPreflight{}, &GoogleGroupAlreadyBoundError{Binding: bound}
		}
	}
	if !preflight.BindingChanged {
		return preflight, nil
	}
	sameGoogleGroup := preflight.PreviousGroup != nil &&
		preflight.PreviousGroup.Exists &&
		preflight.RequestedGroup != nil &&
		preflight.RequestedGroup.Exists &&
		preflight.PreviousGroup.ID == preflight.RequestedGroup.ID
	preflight.RequiredConfirmation.DeletePreviousGroup = preflight.PreviousGroup != nil &&
		preflight.PreviousGroup.Exists &&
		!sameGoogleGroup
	preflight.RequiredConfirmation.OverwriteRequestedGroupMembers = preflight.RequestedGroup != nil &&
		preflight.RequestedGroup.Exists &&
		len(preflight.RequestedGroup.Members) > 0 &&
		!sameGoogleGroup
	return preflight, nil
}

type coreGroup struct {
	Name string `json:"name"`
}

func getCoreGroup(groupID string) (coreGroup, error) {
	var group coreGroup
	if err := sentinel.Get("/api/groups/"+groupID, &group); err != nil {
		return coreGroup{}, err
	}
	if strings.TrimSpace(group.Name) == "" {
		return coreGroup{}, errors.New("sentinel group has no name")
	}
	return group, nil
}

func QueueGoogleBinding(
	ctx context.Context,
	groupID string,
	requestedEmail string,
	expectedCurrentBindingID string,
	confirmations GoogleBindingConfirmations,
) (*model.GroupGoogleBinding, GoogleBindingPreflight, error) {
	bindingMutationMu.Lock()
	defer bindingMutationMu.Unlock()

	preflight, err := PreflightGoogleBinding(ctx, groupID, requestedEmail)
	if err != nil {
		return nil, GoogleBindingPreflight{}, err
	}
	if preflight.CurrentBinding != nil && preflight.CurrentBinding.Status == BindingStatusPending {
		return nil, preflight, ErrBindingSyncInProgress
	}
	currentBindingID := ""
	if preflight.CurrentBinding != nil {
		currentBindingID = preflight.CurrentBinding.ID
	}
	if currentBindingID != expectedCurrentBindingID {
		return nil, preflight, &BindingStateChangedError{Preflight: preflight}
	}
	if (preflight.RequiredConfirmation.OverwriteRequestedGroupMembers && !confirmations.OverwriteRequestedGroupMembers) ||
		(preflight.RequiredConfirmation.DeletePreviousGroup && !confirmations.DeletePreviousGroup) {
		return nil, preflight, &ConfirmationRequiredError{Preflight: preflight}
	}
	if !preflight.BindingChanged {
		return preflight.CurrentBinding, preflight, nil
	}

	canonicalRequestedEmail := preflight.RequestedEmail
	if preflight.RequestedGroup != nil {
		if preflight.RequestedGroup.Exists {
			canonicalRequestedEmail = preflight.RequestedGroup.Email
		}
	}
	operationID := ulid.Make().Prefixed("gbo")
	previousGoogleGroupEmail := ""
	previousGoogleGroupID := ""
	if preflight.RequiredConfirmation.DeletePreviousGroup {
		previousGoogleGroupEmail = preflight.PreviousGroup.Email
		previousGoogleGroupID = preflight.PreviousGroup.ID
	}
	targetGoogleGroupID := ""
	if preflight.RequestedGroup != nil && preflight.RequestedGroup.Exists {
		targetGoogleGroupID = preflight.RequestedGroup.ID
	}

	var binding model.GroupGoogleBinding
	if preflight.CurrentBinding != nil {
		binding = *preflight.CurrentBinding
		updates := map[string]any{
			"status":                      BindingStatusPending,
			"operation_id":                operationID,
			"previous_google_group_email": previousGoogleGroupEmail,
			"previous_google_group_id":    previousGoogleGroupID,
			"target_google_group_id":      targetGoogleGroupID,
			"delete_requested":            canonicalRequestedEmail == "",
			"allow_bulk_removals":         confirmations.OverwriteRequestedGroupMembers,
			"last_sync_error":             "",
		}
		if canonicalRequestedEmail != "" {
			updates["google_group_email"] = canonicalRequestedEmail
		}
		result := database.DB.Model(&model.GroupGoogleBinding{}).
			Where("id = ? AND group_id = ? AND status = ?", binding.ID, groupID, BindingStatusActive).
			Updates(updates)
		if result.Error != nil {
			return nil, preflight, result.Error
		}
		if result.RowsAffected != 1 {
			return nil, preflight, &BindingStateChangedError{Preflight: preflight}
		}
		if err := database.DB.First(&binding, "id = ?", binding.ID).Error; err != nil {
			return nil, preflight, err
		}
	} else {
		if canonicalRequestedEmail == "" {
			return nil, preflight, nil
		}
		binding = model.GroupGoogleBinding{
			ID:                       ulid.Make().Prefixed("ggb"),
			GroupID:                  groupID,
			GoogleGroupEmail:         canonicalRequestedEmail,
			Status:                   BindingStatusPending,
			OperationID:              operationID,
			PreviousGoogleGroupEmail: previousGoogleGroupEmail,
			PreviousGoogleGroupID:    previousGoogleGroupID,
			TargetGoogleGroupID:      targetGoogleGroupID,
			AllowBulkRemovals:        confirmations.OverwriteRequestedGroupMembers,
		}
		if err := database.DB.Create(&binding).Error; err != nil {
			return nil, preflight, err
		}
	}
	logger.SugarLogger.Infof(
		"google binding: queued operation=%s group=%s google=%s delete=%t",
		binding.OperationID,
		binding.GroupID,
		binding.GoogleGroupEmail,
		binding.DeleteRequested,
	)
	TriggerReconcile()
	return &binding, preflight, nil
}

func applyPendingGoogleBinding(ctx context.Context, binding model.GroupGoogleBinding) error {
	if binding.Status != BindingStatusPending || binding.OperationID == "" {
		return nil
	}
	if binding.DeleteRequested {
		groupKey := binding.PreviousGoogleGroupID
		if groupKey == "" {
			groupKey = binding.GoogleGroupEmail
		}
		if err := deleteGoogleGroup(ctx, groupKey); err != nil {
			return err
		}
		result := database.DB.Where(
			"id = ? AND operation_id = ? AND status = ?",
			binding.ID,
			binding.OperationID,
			BindingStatusPending,
		).Delete(&model.GroupGoogleBinding{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrBindingSyncInProgress
		}
		return nil
	}

	group, exists, err := getGoogleGroup(ctx, binding.GoogleGroupEmail)
	if err != nil {
		return err
	}
	if exists && binding.TargetGoogleGroupID != "" && group.Id != binding.TargetGoogleGroupID {
		return ErrGoogleGroupStateChanged
	}
	if !exists {
		if binding.TargetGoogleGroupID != "" {
			return ErrGoogleGroupStateChanged
		}
		coreGroup, err := getCoreGroup(binding.GroupID)
		if err != nil {
			return fmt.Errorf("load sentinel group: %w", err)
		}
		group, err = createGoogleGroup(ctx, binding.GoogleGroupEmail, coreGroup.Name)
		if err != nil {
			return err
		}
		binding.TargetGoogleGroupID = group.Id
		result := database.DB.Model(&model.GroupGoogleBinding{}).
			Where(
				"id = ? AND operation_id = ? AND status = ?",
				binding.ID,
				binding.OperationID,
				BindingStatusPending,
			).
			Update("target_google_group_id", binding.TargetGoogleGroupID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrBindingSyncInProgress
		}
	}

	if err := reconcileBinding(ctx, binding, binding.AllowBulkRemovals); err != nil {
		return err
	}
	if binding.PreviousGoogleGroupEmail != "" && !strings.EqualFold(binding.PreviousGoogleGroupEmail, binding.GoogleGroupEmail) {
		current, err := pendingGoogleBinding(binding.ID, binding.OperationID)
		if err != nil {
			return err
		}
		if !current {
			return ErrBindingSyncInProgress
		}
		groupKey := binding.PreviousGoogleGroupID
		if groupKey == "" {
			groupKey = binding.PreviousGoogleGroupEmail
		}
		if err := deleteGoogleGroup(ctx, groupKey); err != nil {
			return err
		}
	}

	result := database.DB.Model(&model.GroupGoogleBinding{}).
		Where(
			"id = ? AND operation_id = ? AND status = ?",
			binding.ID,
			binding.OperationID,
			BindingStatusPending,
		).
		Updates(map[string]any{
			"status":                      BindingStatusActive,
			"operation_id":                "",
			"previous_google_group_email": "",
			"previous_google_group_id":    "",
			"target_google_group_id":      "",
			"delete_requested":            false,
			"allow_bulk_removals":         false,
			"last_sync_error":             "",
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrBindingSyncInProgress
	}
	return nil
}

func pendingGoogleBinding(bindingID string, operationID string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.GroupGoogleBinding{}).
		Where(
			"id = ? AND operation_id = ? AND status = ?",
			bindingID,
			operationID,
			BindingStatusPending,
		).
		Count(&count).Error
	return count == 1, err
}

func recordGoogleBindingSyncError(binding model.GroupGoogleBinding, syncErr error) {
	message := syncErr.Error()
	if len(message) > 4096 {
		message = message[:4096]
	}
	if err := database.DB.Model(&model.GroupGoogleBinding{}).
		Where(
			"id = ? AND operation_id = ? AND status = ?",
			binding.ID,
			binding.OperationID,
			BindingStatusPending,
		).
		Update("last_sync_error", message).Error; err != nil {
		logger.SugarLogger.Errorf("google sync: record pending binding error: %v", err)
	}
}
