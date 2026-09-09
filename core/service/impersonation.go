package service

import (
	"errors"
	"fmt"

	"github.com/gaucho-racing/sentinel/core/model"
	"gorm.io/gorm"
)

const ImpersonationScope = "sentinel:impersonate"
const ImpersonationTokenTTLSeconds = 5 * 60
const defaultImpersonationScope = "user:read groups:read"
const sentinelImpersonationScope = "sentinel:all"
const sentinelClientID = "sentinel"

var ErrImpersonationCallerNotServiceAccount = errors.New("impersonation caller must be a service account")
var ErrImpersonationUserNotFound = errors.New("impersonation user not found")
var ErrImpersonationApplicationNotFound = errors.New("impersonation application not found")
var ErrImpersonationAccessDenied = errors.New("user does not have access to the application")

type ImpersonationResult struct {
	AccessToken   string
	TokenID       string
	UserID        string
	ApplicationID string
	ClientID      string
	Scope         string
	ExpiresIn     int
}

func ImpersonateUser(actorEntityID string, userID string, applicationID string) (ImpersonationResult, error) {
	actor, err := GetServiceAccountByEntityID(actorEntityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ImpersonationResult{}, ErrImpersonationCallerNotServiceAccount
		}
		return ImpersonationResult{}, fmt.Errorf("load impersonation caller: %w", err)
	}

	user, err := GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ImpersonationResult{}, ErrImpersonationUserNotFound
		}
		return ImpersonationResult{}, fmt.Errorf("load impersonation user: %w", err)
	}

	application, err := GetApplicationByID(applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ImpersonationResult{}, ErrImpersonationApplicationNotFound
		}
		return ImpersonationResult{}, fmt.Errorf("load impersonation application: %w", err)
	}

	groups, err := impersonationGroups(user.EntityID, application)
	if err != nil {
		return ImpersonationResult{}, err
	}

	scope := defaultImpersonationScope
	if application.ClientID == sentinelClientID {
		scope = sentinelImpersonationScope
	}

	groupNames := make([]string, 0, len(groups))
	groupIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
		groupIDs = append(groupIDs, group.ID)
	}

	claims := map[string]interface{}{
		"entity_type": string(model.EntityTypeUser),
		"user_id":     user.ID,
		"groups":      groupNames,
		"group_ids":   groupIDs,
		"act": map[string]interface{}{
			"sub":                actorEntityID,
			"service_account_id": actor.ID,
		},
	}

	accessToken, tokenID, err := GenerateTokenWithActor(
		user.EntityID,
		application.ClientID,
		scope,
		ImpersonationTokenTTLSeconds,
		claims,
		actorEntityID,
	)
	if err != nil {
		return ImpersonationResult{}, fmt.Errorf("mint impersonation token: %w", err)
	}

	return ImpersonationResult{
		AccessToken:   accessToken,
		TokenID:       tokenID,
		UserID:        user.ID,
		ApplicationID: application.ID,
		ClientID:      application.ClientID,
		Scope:         scope,
		ExpiresIn:     ImpersonationTokenTTLSeconds,
	}, nil
}

func impersonationGroups(entityID string, application model.Application) ([]model.Group, error) {
	userGroups, err := GetGroupsForEntity(entityID)
	if err != nil {
		return nil, fmt.Errorf("load user groups: %w", err)
	}

	applicationGroups, err := GetGroupsForApplication(application.ID)
	if err != nil {
		return nil, fmt.Errorf("load application groups: %w", err)
	}
	if !passesApplicationGate(userGroups, applicationGroups) {
		return nil, ErrImpersonationAccessDenied
	}
	if application.ClientID == sentinelClientID {
		return userGroups, nil
	}

	sentinelApplication, err := GetApplicationByClientID(sentinelClientID)
	if err != nil {
		return nil, fmt.Errorf("load sentinel application: %w", err)
	}
	sentinelGroups, err := GetGroupsForApplication(sentinelApplication.ID)
	if err != nil {
		return nil, fmt.Errorf("load sentinel groups: %w", err)
	}

	allowed := make(map[string]struct{}, len(applicationGroups)+len(sentinelGroups))
	for _, group := range applicationGroups {
		allowed[group.ID] = struct{}{}
	}
	for _, group := range sentinelGroups {
		allowed[group.ID] = struct{}{}
	}

	filtered := make([]model.Group, 0, len(userGroups))
	for _, group := range userGroups {
		if _, ok := allowed[group.ID]; ok {
			filtered = append(filtered, group)
		}
	}
	return filtered, nil
}

func passesApplicationGate(userGroups []model.Group, applicationGroups []GroupWithRequired) bool {
	userGroupIDs := make(map[string]struct{}, len(userGroups))
	for _, group := range userGroups {
		userGroupIDs[group.ID] = struct{}{}
	}

	hasRequiredGroup := false
	for _, group := range applicationGroups {
		if !group.Required {
			continue
		}
		hasRequiredGroup = true
		if _, ok := userGroupIDs[group.ID]; ok {
			return true
		}
	}
	return !hasRequiredGroup
}
