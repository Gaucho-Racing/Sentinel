package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gaucho-racing/sentinel/google/database"
	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/sentinel/google/pkg/sentinel"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

var ErrBindingNotFound = errors.New("group google binding not found")
var ErrGoogleSyncUnavailable = errors.New("google group management is not configured")

const ManagedGoogleGroupOwnerEmail = "team@gauchoracing.com"

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

func CreateGoogleBinding(binding model.GroupGoogleBinding) (model.GroupGoogleBinding, error) {
	if binding.ID == "" {
		binding.ID = ulid.Make().Prefixed("ggb")
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		return model.GroupGoogleBinding{}, err
	}
	return binding, nil
}

func updateGoogleBinding(binding model.GroupGoogleBinding, googleGroupEmail string) (model.GroupGoogleBinding, error) {
	binding.GoogleGroupEmail = googleGroupEmail
	if err := database.DB.Save(&binding).Error; err != nil {
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

func ApplyGoogleBinding(
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
	createdRequestedGroup := false
	if preflight.RequestedGroup != nil {
		if preflight.RequestedGroup.Exists {
			canonicalRequestedEmail = preflight.RequestedGroup.Email
		} else {
			group, err := getCoreGroup(groupID)
			if err != nil {
				return nil, preflight, fmt.Errorf("load sentinel group: %w", err)
			}
			created, err := createGoogleGroup(ctx, preflight.RequestedEmail, group.Name)
			if err != nil {
				return nil, preflight, err
			}
			createdRequestedGroup = true
			canonicalRequestedEmail = strings.ToLower(created.Email)
		}
		candidate := model.GroupGoogleBinding{
			GroupID:          groupID,
			GoogleGroupEmail: canonicalRequestedEmail,
		}
		if err := reconcileBinding(
			ctx,
			candidate,
			confirmations.OverwriteRequestedGroupMembers,
		); err != nil {
			if createdRequestedGroup {
				_ = deleteGoogleGroup(ctx, canonicalRequestedEmail)
			}
			return nil, preflight, err
		}
	}

	if preflight.RequiredConfirmation.DeletePreviousGroup {
		if err := deleteGoogleGroup(ctx, preflight.PreviousGroup.Email); err != nil {
			return nil, preflight, err
		}
	}

	if canonicalRequestedEmail == "" {
		if preflight.CurrentBinding != nil {
			if err := DeleteGoogleBinding(groupID, preflight.CurrentBinding.ID); err != nil {
				return nil, preflight, err
			}
		}
		return nil, preflight, nil
	}
	if preflight.CurrentBinding != nil {
		binding, err := updateGoogleBinding(*preflight.CurrentBinding, canonicalRequestedEmail)
		if err != nil {
			return nil, preflight, err
		}
		return &binding, preflight, nil
	}
	binding, err := CreateGoogleBinding(model.GroupGoogleBinding{
		GroupID:          groupID,
		GoogleGroupEmail: canonicalRequestedEmail,
	})
	if err != nil {
		return nil, preflight, err
	}
	return &binding, preflight, nil
}

// DeleteGoogleBinding scopes the delete to (groupID, bindingID) so a tampered
// request can't drop a binding for a different group.
func DeleteGoogleBinding(groupID, bindingID string) error {
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, bindingID).Delete(&model.GroupGoogleBinding{}).Error; err != nil {
		return err
	}
	return nil
}
