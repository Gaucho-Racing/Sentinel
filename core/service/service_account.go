package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

// ServiceAccountAllowedScopes is the closed set of scopes an admin can
// mint onto a service account's bearer JWT. Deliberately read-only:
//   - sentinel:all is first-party only (never issued to third-party
//     OAuth flows; SAs are arms-length automations of an app)
//   - openid/profile/email/offline_access are OIDC user-identity scopes,
//     meaningless for a non-human SA token
//   - *:write scopes are excluded by design — SA-driven mutations should
//     go through human-authed flows so there's accountability
var ServiceAccountAllowedScopes = []string{
	"user:read",
	"groups:read",
	"applications:read",
}

// ErrInvalidServiceAccountScope is returned by ValidateServiceAccountScope
// when the requested scope contains any scope outside the allow-list.
var ErrInvalidServiceAccountScope = errors.New("scope contains a value not allowed for service accounts")

// ValidateServiceAccountScope checks every space-separated scope in `s`
// against ServiceAccountAllowedScopes. Empty scope is allowed (a token
// with no scope can only be used for endpoints that don't require any).
func ValidateServiceAccountScope(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	allowed := make(map[string]struct{}, len(ServiceAccountAllowedScopes))
	for _, a := range ServiceAccountAllowedScopes {
		allowed[a] = struct{}{}
	}
	for _, scope := range strings.Fields(s) {
		if _, ok := allowed[scope]; !ok {
			return fmt.Errorf("%w: %q", ErrInvalidServiceAccountScope, scope)
		}
	}
	return nil
}

// SAMaxTokenLifetimeSeconds is the exp-claim duration used when an admin
// picks "never expires" for an SA. Set to ~100 years so the token is
// effectively non-expiring in any operationally meaningful window. The
// frontend renders dates this far in the future as "Never" rather than
// surfacing the synthetic year.
const SAMaxTokenLifetimeSeconds = 100 * 365 * 24 * 60 * 60

func GetAllServiceAccounts() ([]model.ServiceAccount, error) {
	var serviceAccounts []model.ServiceAccount
	if err := database.DB.Find(&serviceAccounts).Error; err != nil {
		return []model.ServiceAccount{}, err
	}
	for i := range serviceAccounts {
		PopulateServiceAccount(&serviceAccounts[i])
	}
	return serviceAccounts, nil
}

func GetServiceAccountByID(id string) (model.ServiceAccount, error) {
	var sa model.ServiceAccount
	if err := database.DB.Where("id = ?", id).First(&sa).Error; err != nil {
		return model.ServiceAccount{}, err
	}
	PopulateServiceAccount(&sa)
	return sa, nil
}

func GetServiceAccountByEntityID(entityID string) (model.ServiceAccount, error) {
	var sa model.ServiceAccount
	if err := database.DB.Where("entity_id = ?", entityID).First(&sa).Error; err != nil {
		return model.ServiceAccount{}, err
	}
	PopulateServiceAccount(&sa)
	return sa, nil
}

func GetServiceAccountByName(name string) (model.ServiceAccount, error) {
	var sa model.ServiceAccount
	if err := database.DB.Where("name = ?", name).First(&sa).Error; err != nil {
		return model.ServiceAccount{}, err
	}
	PopulateServiceAccount(&sa)
	return sa, nil
}

func GetServiceAccountsByApplicationID(applicationID string) ([]model.ServiceAccount, error) {
	var serviceAccounts []model.ServiceAccount
	if err := database.DB.Where("application_id = ?", applicationID).Find(&serviceAccounts).Error; err != nil {
		return []model.ServiceAccount{}, err
	}
	for i := range serviceAccounts {
		PopulateServiceAccount(&serviceAccounts[i])
	}
	return serviceAccounts, nil
}

func CreateServiceAccount(sa model.ServiceAccount) (model.ServiceAccount, error) {
	if sa.ID == "" {
		sa.ID = ulid.Make().Prefixed("sa")
	}
	if err := database.DB.Create(&sa).Error; err != nil {
		return model.ServiceAccount{}, err
	}
	PopulateServiceAccount(&sa)
	return sa, nil
}

func UpdateServiceAccount(sa model.ServiceAccount) (model.ServiceAccount, error) {
	if err := database.DB.Save(&sa).Error; err != nil {
		return model.ServiceAccount{}, err
	}
	PopulateServiceAccount(&sa)
	return sa, nil
}

func DeleteServiceAccount(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.ServiceAccount{}).Error; err != nil {
		return err
	}
	return nil
}

func PopulateServiceAccount(sa *model.ServiceAccount) {
	groups, err := GetGroupsForEntity(sa.EntityID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get groups for service account %s: %v", sa.ID, err)
	}
	sa.Groups = groups

	token, err := GetLatestTokenForEntity(sa.EntityID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.SugarLogger.Errorf("Failed to get active token for service account %s: %v", sa.ID, err)
		}
		sa.ActiveToken = nil
		return
	}
	sa.ActiveToken = &token
}

// CreateServiceAccountForApp is the all-in-one path the HTTP API uses to
// stand up a new SA: mint an Entity (type SERVICE_ACCOUNT) and link a
// ServiceAccount row to it. Failure on the second step leaves an
// orphaned entity, which is harmless (no auth path resolves it).
func CreateServiceAccountForApp(applicationID, name, scope string, ttlDays int, createdBy string) (model.ServiceAccount, error) {
	entity, err := CreateEntity(model.Entity{
		Type: model.EntityTypeServiceAccount,
	})
	if err != nil {
		return model.ServiceAccount{}, err
	}
	sa, err := CreateServiceAccount(model.ServiceAccount{
		EntityID:      entity.ID,
		ApplicationID: applicationID,
		Name:          name,
		Scope:         scope,
		TTLDays:       ttlDays,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return model.ServiceAccount{}, err
	}
	return sa, nil
}

// MintServiceAccountToken delete-and-mints the SA's bearer JWT. Always
// revokes any existing token rows for the SA's entity first so a SA only
// ever has one active credential. Returns the raw signed JWT alongside
// the persisted Token row — the JWT is exposed only at this boundary
// (create + rotate); subsequent reads never include it.
//
// scope must already have passed ValidateServiceAccountScope. ttlDays
// must be one of: 30, 90, 365, 0 (=never). The auth_token row carries
// the real exp; ttlDays is also persisted on the SA so rotation can
// reuse the admin's chosen window.
func MintServiceAccountToken(sa model.ServiceAccount) (model.Token, string, error) {
	app, err := GetApplicationByID(sa.ApplicationID)
	if err != nil {
		return model.Token{}, "", fmt.Errorf("load application: %w", err)
	}

	// Revoke any existing rows first. One SA = one token at a time.
	if err := DeleteTokensForEntity(sa.EntityID); err != nil {
		return model.Token{}, "", fmt.Errorf("revoke existing tokens: %w", err)
	}

	ttlSeconds := sa.TTLDays * 24 * 60 * 60
	if sa.TTLDays == 0 {
		ttlSeconds = SAMaxTokenLifetimeSeconds
	}

	// Custom claims carry SA-identifying metadata so token consumers can
	// distinguish a service-account token from a user JWT without
	// re-resolving the subject entity. `groups` claim is intentionally
	// omitted — downstream services that need the SA's groups should
	// resolve them dynamically via /entities/:id/groups so a group
	// membership change takes effect on the next request rather than
	// waiting for the SA to rotate.
	claims := map[string]any{
		"type":  "service_account",
		"sa_id": sa.ID,
	}

	raw, tokenID, err := GenerateToken(sa.EntityID, app.ClientID, sa.Scope, ttlSeconds, claims)
	if err != nil {
		return model.Token{}, "", err
	}

	var token model.Token
	if err := database.DB.Where("id = ?", tokenID).First(&token).Error; err != nil {
		return model.Token{}, "", fmt.Errorf("read back minted token: %w", err)
	}
	return token, raw, nil
}
