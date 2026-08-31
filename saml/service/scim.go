package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
	"gorm.io/gorm"
)

var (
	ErrInvalidSCIMConfiguration = errors.New("invalid SCIM configuration")
	ErrSCIMSyncInProgress       = errors.New("a SCIM synchronization is already running for this application")
)

type SCIMConfigurationView struct {
	ApplicationID   string                 `json:"application_id"`
	Endpoint        string                 `json:"endpoint"`
	TokenConfigured bool                   `json:"token_configured"`
	TokenExpiresAt  *time.Time             `json:"token_expires_at"`
	Enabled         bool                   `json:"enabled"`
	SyncInterval    model.SCIMSyncInterval `json:"sync_interval"`
	NextSyncAt      *time.Time             `json:"next_sync_at"`
	LastSyncAt      *time.Time             `json:"last_sync_at"`
	LastSyncStatus  model.SCIMSyncStatus   `json:"last_sync_status"`
	LastSyncError   string                 `json:"last_sync_error"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedAt       time.Time              `json:"created_at"`
}

type ProvisioningUser struct {
	EntityID  string `json:"entity_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ProvisioningGroup struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type ProvisioningSnapshot struct {
	Users  []ProvisioningUser  `json:"users"`
	Groups []ProvisioningGroup `json:"groups"`
}

type SCIMSkippedUser = model.SCIMSkippedUser

type SCIMSyncResult struct {
	DryRun             bool              `json:"dry_run"`
	DesiredUsers       int               `json:"desired_users"`
	DesiredGroups      int               `json:"desired_groups"`
	UsersCreated       int               `json:"users_created"`
	UsersUpdated       int               `json:"users_updated"`
	UsersDeactivated   int               `json:"users_deactivated"`
	GroupsCreated      int               `json:"groups_created"`
	GroupsUpdated      int               `json:"groups_updated"`
	MembershipsAdded   int               `json:"memberships_added"`
	MembershipsRemoved int               `json:"memberships_removed"`
	SkippedUsers       []SCIMSkippedUser `json:"skipped_users"`
	ValidationErrors   []string          `json:"validation_errors"`
	CompletedAt        time.Time         `json:"completed_at"`
}

func GetSCIMConfiguration(applicationID string) (model.SCIMConfiguration, error) {
	var configuration model.SCIMConfiguration
	err := database.DB.Where("application_id = ?", applicationID).First(&configuration).Error
	return configuration, err
}

func SCIMConfigurationResponse(configuration model.SCIMConfiguration) SCIMConfigurationView {
	if _, err := scimSyncIntervalDuration(configuration.SyncInterval); err != nil {
		configuration.SyncInterval = scimDefaultSyncInterval
	}
	return SCIMConfigurationView{
		ApplicationID:   configuration.ApplicationID,
		Endpoint:        configuration.Endpoint,
		TokenConfigured: configuration.AccessToken != "",
		TokenExpiresAt:  configuration.TokenExpiresAt,
		Enabled:         configuration.Enabled,
		SyncInterval:    configuration.SyncInterval,
		NextSyncAt:      configuration.NextSyncAt,
		LastSyncAt:      configuration.LastSyncAt,
		LastSyncStatus:  configuration.LastSyncStatus,
		LastSyncError:   configuration.LastSyncError,
		UpdatedAt:       configuration.UpdatedAt,
		CreatedAt:       configuration.CreatedAt,
	}
}

func UpsertSCIMConfiguration(applicationID, endpoint, accessToken string, tokenExpiresAt *time.Time, enabled bool, syncInterval model.SCIMSyncInterval) (model.SCIMConfiguration, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if err := validateSCIMEndpoint(endpoint); err != nil {
		return model.SCIMConfiguration{}, err
	}
	if syncInterval == "" {
		syncInterval = scimDefaultSyncInterval
	}
	intervalDuration, err := scimSyncIntervalDuration(syncInterval)
	if err != nil {
		return model.SCIMConfiguration{}, err
	}
	sp, err := GetServiceProvider(applicationID)
	if err != nil {
		return model.SCIMConfiguration{}, fmt.Errorf("load SAML configuration: %w", err)
	}
	if sp.Profile != model.ProfileAWSIdentityCenter {
		return model.SCIMConfiguration{}, fmt.Errorf("%w: SCIM provisioning requires the AWS IAM Identity Center SAML profile", ErrInvalidSCIMConfiguration)
	}

	configuration, err := GetSCIMConfiguration(applicationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SCIMConfiguration{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		configuration = model.SCIMConfiguration{
			ApplicationID:  applicationID,
			LastSyncStatus: model.SCIMSyncStatusNever,
		}
	}
	if strings.TrimSpace(accessToken) != "" {
		configuration.AccessToken = strings.TrimSpace(accessToken)
	}
	if configuration.AccessToken == "" {
		return model.SCIMConfiguration{}, fmt.Errorf("%w: access token is required", ErrInvalidSCIMConfiguration)
	}
	if tokenExpiresAt != nil && !tokenExpiresAt.After(time.Now()) {
		return model.SCIMConfiguration{}, fmt.Errorf("%w: token expiration must be in the future", ErrInvalidSCIMConfiguration)
	}
	now := time.Now().UTC()
	intervalChanged := configuration.SyncInterval != syncInterval
	wasEnabled := configuration.Enabled
	configuration.Endpoint = endpoint
	configuration.TokenExpiresAt = tokenExpiresAt
	configuration.Enabled = enabled
	configuration.SyncInterval = syncInterval
	if !enabled {
		configuration.NextSyncAt = nil
	} else if !wasEnabled || intervalChanged || configuration.NextSyncAt == nil {
		nextSyncAt := now.Add(intervalDuration)
		configuration.NextSyncAt = &nextSyncAt
	}
	if err := database.DB.Save(&configuration).Error; err != nil {
		return model.SCIMConfiguration{}, err
	}
	return configuration, nil
}

func DeleteSCIMConfiguration(applicationID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", "sentinel-scim-queue:"+applicationID).Error; err != nil {
			return err
		}
		var activeRuns int64
		if err := tx.Model(&model.SCIMSyncRun{}).Where("application_id = ? AND status IN ?", applicationID, []model.SCIMSyncRunStatus{
			model.SCIMSyncRunStatusQueued,
			model.SCIMSyncRunStatusRunning,
		}).Count(&activeRuns).Error; err != nil {
			return err
		}
		if activeRuns > 0 {
			return ErrSCIMSyncInProgress
		}
		result := tx.Where("application_id = ?", applicationID).Delete(&model.SCIMConfiguration{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("application_id = ?", applicationID).Delete(&model.SCIMResource{}).Error
	})
}

func TestSCIMConfiguration(ctx context.Context, applicationID string) error {
	if err := requireAWSProfile(applicationID); err != nil {
		return err
	}
	configuration, err := GetSCIMConfiguration(applicationID)
	if err != nil {
		return err
	}
	return NewSCIMClient(configuration.Endpoint, configuration.AccessToken).Test(ctx)
}

func PreviewSCIMSync(applicationID string) (SCIMSyncResult, error) {
	if err := requireAWSProfile(applicationID); err != nil {
		return SCIMSyncResult{}, err
	}
	snapshot, err := loadProvisioningSnapshot(applicationID)
	if err != nil {
		return SCIMSyncResult{}, err
	}
	snapshot, skippedUsers := prepareProvisioningSnapshot(snapshot)
	return SCIMSyncResult{
		DryRun:           true,
		DesiredUsers:     len(snapshot.Users),
		DesiredGroups:    len(snapshot.Groups),
		SkippedUsers:     skippedUsers,
		ValidationErrors: validateProvisioningSnapshot(snapshot),
		CompletedAt:      time.Now().UTC(),
	}, nil
}

func executeSCIMSync(ctx context.Context, applicationID string) (SCIMSyncResult, error) {
	if err := requireAWSProfile(applicationID); err != nil {
		return SCIMSyncResult{}, err
	}
	configuration, err := GetSCIMConfiguration(applicationID)
	if err != nil {
		return SCIMSyncResult{}, err
	}
	if !configuration.Enabled {
		return SCIMSyncResult{}, fmt.Errorf("%w: provisioning is disabled", ErrInvalidSCIMConfiguration)
	}
	if configuration.TokenExpiresAt != nil && !configuration.TokenExpiresAt.After(time.Now()) {
		return SCIMSyncResult{}, fmt.Errorf("%w: the SCIM access token has expired", ErrInvalidSCIMConfiguration)
	}

	snapshot, err := loadProvisioningSnapshot(applicationID)
	if err != nil {
		return SCIMSyncResult{}, err
	}
	snapshot, skippedUsers := prepareProvisioningSnapshot(snapshot)
	result := SCIMSyncResult{
		DesiredUsers:     len(snapshot.Users),
		DesiredGroups:    len(snapshot.Groups),
		SkippedUsers:     skippedUsers,
		ValidationErrors: validateProvisioningSnapshot(snapshot),
	}
	if len(result.ValidationErrors) > 0 {
		return result, fmt.Errorf("provisioning scope contains invalid users")
	}

	return withSCIMApplicationLock(ctx, applicationID, func() (SCIMSyncResult, error) {
		for _, skipped := range result.SkippedUsers {
			logger.SugarLogger.Warnw(
				"SCIM synchronization skipped user",
				"application_id", applicationID,
				"entity_id", skipped.EntityID,
				"username", skipped.Username,
				"groups", skipped.Groups,
				"reason", skipped.Reason,
			)
		}

		result, syncErr := reconcileSCIM(ctx, configuration, snapshot, result)
		result.CompletedAt = time.Now().UTC()
		return result, syncErr
	})
}

func reconcileSCIM(ctx context.Context, configuration model.SCIMConfiguration, snapshot ProvisioningSnapshot, result SCIMSyncResult) (SCIMSyncResult, error) {
	client := NewSCIMClient(configuration.Endpoint, configuration.AccessToken)
	userProviderIDs := make(map[string]string, len(snapshot.Users))
	desiredUserIDs := make(map[string]bool, len(snapshot.Users))

	for _, user := range snapshot.Users {
		desired := provisioningSCIMUser(user)
		remote, err := client.FindUser(ctx, "externalId", user.EntityID)
		if err != nil {
			return result, fmt.Errorf("find user %s: %w", user.Email, err)
		}
		if remote == nil {
			remote, err = client.FindUser(ctx, "userName", desired.UserName)
			if err != nil {
				return result, fmt.Errorf("find existing user %s: %w", user.Email, err)
			}
		}
		if remote == nil {
			created, err := client.CreateUser(ctx, desired)
			if err != nil {
				return result, fmt.Errorf("create user %s: %w", user.Email, err)
			}
			remote = &created
			result.UsersCreated++
		} else if !scimUserMatches(*remote, desired) {
			if err := client.UpdateUser(ctx, remote.ID, desired); err != nil {
				return result, fmt.Errorf("update user %s: %w", user.Email, err)
			}
			result.UsersUpdated++
		}
		if remote.ID == "" {
			return result, fmt.Errorf("SCIM user %s has no provider ID", user.Email)
		}
		desiredUserIDs[user.EntityID] = true
		userProviderIDs[user.EntityID] = remote.ID
		if err := saveSCIMResource(configuration.ApplicationID, model.SCIMResourceTypeUser, user.EntityID, remote.ID); err != nil {
			return result, err
		}
	}

	var managedUsers []model.SCIMResource
	if err := database.DB.Where("application_id = ? AND resource_type = ?", configuration.ApplicationID, model.SCIMResourceTypeUser).Find(&managedUsers).Error; err != nil {
		return result, err
	}
	for _, managed := range managedUsers {
		if desiredUserIDs[managed.SentinelID] {
			continue
		}
		if err := client.SetUserActive(ctx, managed.ProviderID, false); err != nil {
			return result, fmt.Errorf("deactivate user %s: %w", managed.SentinelID, err)
		}
		if err := database.DB.Delete(&managed).Error; err != nil {
			return result, err
		}
		result.UsersDeactivated++
	}

	desiredGroupIDs := make(map[string]bool, len(snapshot.Groups))
	for _, group := range snapshot.Groups {
		desiredGroupIDs[group.ID] = true
		remote, err := client.FindGroup(ctx, "externalId", group.ID)
		if err != nil {
			return result, fmt.Errorf("find group %s: %w", group.Name, err)
		}
		if remote == nil {
			remote, err = client.FindGroup(ctx, "displayName", group.Name)
			if err != nil {
				return result, fmt.Errorf("find existing group %s: %w", group.Name, err)
			}
		}
		if remote == nil {
			created, err := client.CreateGroup(ctx, SCIMGroup{ExternalID: group.ID, DisplayName: group.Name})
			if err != nil {
				return result, fmt.Errorf("create group %s: %w", group.Name, err)
			}
			remote = &created
			result.GroupsCreated++
		} else if remote.ExternalID != group.ID || remote.DisplayName != group.Name {
			if err := client.UpdateGroup(ctx, remote.ID, group.ID, group.Name); err != nil {
				return result, fmt.Errorf("update group %s: %w", group.Name, err)
			}
			result.GroupsUpdated++
		}
		if remote.ID == "" {
			return result, fmt.Errorf("SCIM group %s has no provider ID", group.Name)
		}
		if err := saveSCIMResource(configuration.ApplicationID, model.SCIMResourceTypeGroup, group.ID, remote.ID); err != nil {
			return result, err
		}

		currentUsers, err := client.ListUsersForGroup(ctx, remote.ID)
		if err != nil {
			return result, fmt.Errorf("list members for group %s: %w", group.Name, err)
		}
		current := make(map[string]bool, len(currentUsers))
		for _, currentUser := range currentUsers {
			current[currentUser.ID] = true
		}
		desired := make(map[string]bool, len(group.Members))
		for _, entityID := range group.Members {
			if providerID := userProviderIDs[entityID]; providerID != "" {
				desired[providerID] = true
			}
		}
		additions := differenceKeys(desired, current)
		removals := differenceKeys(current, desired)
		if err := client.AddGroupMembers(ctx, remote.ID, additions); err != nil {
			return result, fmt.Errorf("add members to group %s: %w", group.Name, err)
		}
		if err := client.RemoveGroupMembers(ctx, remote.ID, removals); err != nil {
			return result, fmt.Errorf("remove members from group %s: %w", group.Name, err)
		}
		result.MembershipsAdded += len(additions)
		result.MembershipsRemoved += len(removals)
	}

	var managedGroups []model.SCIMResource
	if err := database.DB.Where("application_id = ? AND resource_type = ?", configuration.ApplicationID, model.SCIMResourceTypeGroup).Find(&managedGroups).Error; err != nil {
		return result, err
	}
	for _, managed := range managedGroups {
		if desiredGroupIDs[managed.SentinelID] {
			continue
		}
		currentUsers, err := client.ListUsersForGroup(ctx, managed.ProviderID)
		if err != nil {
			return result, fmt.Errorf("list members for unlinked group %s: %w", managed.SentinelID, err)
		}
		removals := make([]string, 0, len(currentUsers))
		for _, currentUser := range currentUsers {
			removals = append(removals, currentUser.ID)
		}
		if err := client.RemoveGroupMembers(ctx, managed.ProviderID, removals); err != nil {
			return result, fmt.Errorf("remove members from unlinked group %s: %w", managed.SentinelID, err)
		}
		if err := database.DB.Delete(&managed).Error; err != nil {
			return result, err
		}
		result.MembershipsRemoved += len(removals)
	}

	return result, nil
}

func provisioningSCIMUser(user ProvisioningUser) SCIMUser {
	displayName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	email := strings.ToLower(strings.TrimSpace(user.Email))
	return SCIMUser{
		ExternalID:  user.EntityID,
		UserName:    email,
		Name:        SCIMUserName{Formatted: displayName, GivenName: strings.TrimSpace(user.FirstName), FamilyName: strings.TrimSpace(user.LastName)},
		DisplayName: displayName,
		Active:      true,
		Emails:      []SCIMEmail{{Value: email, Type: "work", Primary: true}},
	}
}

func scimUserMatches(remote, desired SCIMUser) bool {
	if remote.ExternalID != desired.ExternalID ||
		!strings.EqualFold(remote.UserName, desired.UserName) ||
		remote.Name.GivenName != desired.Name.GivenName ||
		remote.Name.FamilyName != desired.Name.FamilyName ||
		remote.Name.Formatted != desired.Name.Formatted ||
		remote.DisplayName != desired.DisplayName ||
		remote.Active != desired.Active {
		return false
	}
	for _, email := range remote.Emails {
		if email.Primary && strings.EqualFold(email.Value, desired.Emails[0].Value) {
			return true
		}
	}
	return false
}

func validateProvisioningSnapshot(snapshot ProvisioningSnapshot) []string {
	errorsFound := []string{}
	seenEmails := make(map[string]string)
	for _, user := range snapshot.Users {
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if _, err := mail.ParseAddress(email); err != nil {
			errorsFound = append(errorsFound, fmt.Sprintf("user %s does not have a valid email address", user.EntityID))
		}
		if strings.TrimSpace(user.FirstName) == "" {
			errorsFound = append(errorsFound, fmt.Sprintf("user %s does not have a first name", user.EntityID))
		}
		if strings.TrimSpace(user.LastName) == "" {
			errorsFound = append(errorsFound, fmt.Sprintf("user %s does not have a last name", user.EntityID))
		}
		if previous, exists := seenEmails[email]; email != "" && exists && previous != user.EntityID {
			errorsFound = append(errorsFound, fmt.Sprintf("users %s and %s share email %s", previous, user.EntityID, email))
		}
		seenEmails[email] = user.EntityID
	}
	sort.Strings(errorsFound)
	return errorsFound
}

func prepareProvisioningSnapshot(snapshot ProvisioningSnapshot) (ProvisioningSnapshot, []SCIMSkippedUser) {
	groupsByMember := make(map[string][]string)
	for _, group := range snapshot.Groups {
		for _, memberID := range group.Members {
			groupsByMember[memberID] = append(groupsByMember[memberID], group.Name)
		}
	}

	skippedIDs := make(map[string]bool)
	skippedUsers := make([]SCIMSkippedUser, 0)
	validUsers := make([]ProvisioningUser, 0, len(snapshot.Users))
	for _, user := range snapshot.Users {
		if isValidProvisioningEmail(user.Email) {
			validUsers = append(validUsers, user)
			continue
		}
		groups := groupsByMember[user.EntityID]
		sort.Strings(groups)
		skippedIDs[user.EntityID] = true
		skippedUsers = append(skippedUsers, SCIMSkippedUser{
			EntityID: user.EntityID,
			Username: user.Username,
			Groups:   groups,
			Reason:   "malformed email address",
		})
	}

	groups := make([]ProvisioningGroup, 0, len(snapshot.Groups))
	for _, group := range snapshot.Groups {
		members := make([]string, 0, len(group.Members))
		for _, memberID := range group.Members {
			if !skippedIDs[memberID] {
				members = append(members, memberID)
			}
		}
		group.Members = members
		groups = append(groups, group)
	}
	sort.Slice(skippedUsers, func(i, j int) bool {
		return skippedUsers[i].EntityID < skippedUsers[j].EntityID
	})
	return ProvisioningSnapshot{Users: validUsers, Groups: groups}, skippedUsers
}

func isValidProvisioningEmail(value string) bool {
	email := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email
}

func loadProvisioningSnapshot(applicationID string) (ProvisioningSnapshot, error) {
	var snapshot ProvisioningSnapshot
	route := "/api/core/internal/applications/" + url.PathEscape(applicationID) + "/provisioning-snapshot"
	if err := sentinel.Get(route, &snapshot); err != nil {
		return ProvisioningSnapshot{}, fmt.Errorf("load provisioning scope: %w", err)
	}
	return snapshot, nil
}

func saveSCIMResource(applicationID string, resourceType model.SCIMResourceType, sentinelID, providerID string) error {
	now := time.Now().UTC()
	resource := model.SCIMResource{
		ApplicationID: applicationID,
		ResourceType:  resourceType,
		SentinelID:    sentinelID,
		ProviderID:    providerID,
		LastSyncedAt:  now,
	}
	return database.DB.Save(&resource).Error
}

func differenceKeys(left, right map[string]bool) []string {
	result := []string{}
	for key := range left {
		if !right[key] {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}

func validateSCIMEndpoint(endpoint string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("%w: endpoint must be a valid URL", ErrInvalidSCIMConfiguration)
	}
	hostname := strings.ToLower(parsed.Hostname())
	awsHostname := strings.HasPrefix(hostname, "scim.") &&
		(strings.HasSuffix(hostname, ".amazonaws.com") || strings.HasSuffix(hostname, ".amazonaws.com.cn"))
	if parsed.Scheme != "https" || !awsHostname || parsed.Port() != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: endpoint must be an AWS IAM Identity Center HTTPS SCIM URL without credentials, a port, query string, or fragment", ErrInvalidSCIMConfiguration)
	}
	return nil
}

func requireAWSProfile(applicationID string) error {
	sp, err := GetServiceProvider(applicationID)
	if err != nil {
		return err
	}
	if sp.Profile != model.ProfileAWSIdentityCenter {
		return fmt.Errorf("%w: SCIM provisioning requires the AWS IAM Identity Center SAML profile", ErrInvalidSCIMConfiguration)
	}
	return nil
}

func withSCIMApplicationLock(ctx context.Context, applicationID string, operation func() (SCIMSyncResult, error)) (SCIMSyncResult, error) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return SCIMSyncResult{}, err
	}
	connection, err := sqlDB.Conn(ctx)
	if err != nil {
		return SCIMSyncResult{}, err
	}
	defer connection.Close()

	var locked bool
	if err := connection.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1, 0))", "sentinel-scim:"+applicationID).Scan(&locked); err != nil {
		return SCIMSyncResult{}, err
	}
	if !locked {
		return SCIMSyncResult{}, ErrSCIMSyncInProgress
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(), "SELECT pg_advisory_unlock(hashtextextended($1, 0))", "sentinel-scim:"+applicationID)
	}()
	return operation()
}
