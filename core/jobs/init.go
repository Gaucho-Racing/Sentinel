package jobs

import (
	"errors"
	"strings"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"gorm.io/gorm"
)

const SentinelClientID = "sentinel"
const SentinelApplicationID = "app_01kpy5f8263c4rqnhn9v2akdvf"

const SentinelCoreEntityID = "ent_01kpy5f8263c4rqnhn9y920fkn"
const SentinelCoreServiceAccountID = "sa_01kpy5f8263c4rqnhn9zejxyhk"

// InternalServiceAccountNames is the closed set of pre-seeded SAs that
// non-core services exchange the bootstrap secret for. Each name maps
// 1:1 to a running container in docker-compose (sentinel-discord, etc.)
// so each service can fetch its own token by passing its own name.
//
// The exchange endpoint refuses any name not in this slice — defense
// in depth so a leaked bootstrap secret can't be used to harvest tokens
// for arbitrary (admin-created) service accounts.
var InternalServiceAccountNames = []string{
	"sentinel-discord",
	"sentinel-oauth",
	"sentinel-saml",
	"sentinel-google",
}

// IsInternalServiceAccountName reports whether name is on the
// closed-set internal allowlist. Used to gate the bootstrap exchange.
func IsInternalServiceAccountName(name string) bool {
	for _, n := range InternalServiceAccountNames {
		if n == name {
			return true
		}
	}
	return false
}

func InitializeCore() {
	initializeDefaultApplications()
	initializeDefaultEntities()
	initializeDefaultServiceAccounts()
	initializeAdminsGroup()
	linkAdminsGroupToSentinelApp()
	initializeInternalServiceAccounts()
	logger.SugarLogger.Infoln("Finished initializing sentinel-core")
}

func initializeDefaultApplications() {
	_, err := service.GetApplicationByID(SentinelApplicationID)
	if err == gorm.ErrRecordNotFound {
		app, err := service.CreateApplication(model.Application{
			ID:          SentinelApplicationID,
			Name:        "Sentinel",
			Description: "Gaucho Racing's authentication service",
			ClientID:    SentinelClientID,
			LaunchURL:   "https://sso.gauchoracing.com",
			OwnerID:     SentinelCoreEntityID,
		})
		if err != nil {
			logger.SugarLogger.Fatalf("Failed to create Sentinel application: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Sentinel application (id=%s, client_id=%s)", app.ID, app.ClientID)
		logger.SugarLogger.Infof("Sentinel client secret: %s", app.ClientSecret)

		defaultRedirectURIs := []string{
			"http://localhost:3000/auth/callback",
		}
		for _, uri := range defaultRedirectURIs {
			service.CreateApplicationRedirectURI(app.ID, uri)
		}
		logger.SugarLogger.Infof("Added %d default redirect URIs", len(defaultRedirectURIs))
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Sentinel application: %v", err)
	} else {
		logger.SugarLogger.Infoln("Sentinel application already exists")
	}
}

func initializeDefaultEntities() {
	_, err := service.GetEntityByID(SentinelCoreEntityID)
	if err == gorm.ErrRecordNotFound {
		entity, err := service.CreateEntity(model.Entity{
			ID:   SentinelCoreEntityID,
			Type: model.EntityTypeServiceAccount,
		})
		if err != nil {
			logger.SugarLogger.Fatalf("Failed to create Sentinel core entity: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Sentinel core entity (id=%s)", entity.ID)
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Sentinel core entity: %v", err)
	} else {
		logger.SugarLogger.Infoln("Sentinel core entity already exists")
	}
}

func initializeAdminsGroup() {
	_, err := service.GetGroupByID(service.AdminsGroupID)
	if err == gorm.ErrRecordNotFound {
		if _, err := service.CreateGroup(model.Group{
			ID:             service.AdminsGroupID,
			Name:           "Admins",
			Description:    "Global administrators. Members get owner-equivalent permissions on every group and other admin-gated surfaces.",
			AllowedSources: model.StringSlice{"DIRECT"},
			CreatedBy:      SentinelCoreEntityID,
		}); err != nil {
			logger.SugarLogger.Fatalf("Failed to create Admins group: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Admins group (id=%s)", service.AdminsGroupID)
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Admins group: %v", err)
		return
	}

	if config.AdminEntityIDs == "" {
		return
	}
	for _, raw := range strings.Split(config.AdminEntityIDs, ",") {
		entityID := strings.TrimSpace(raw)
		if entityID == "" {
			continue
		}
		if _, err := service.GetGroupMember(service.AdminsGroupID, entityID); err == nil {
			continue
		}
		if _, err := service.CreateGroupMember(model.GroupMember{
			GroupID:  service.AdminsGroupID,
			EntityID: entityID,
			Source:   string(model.GroupMemberSourceDirect),
			AddedBy:  SentinelCoreEntityID,
		}); err != nil {
			logger.SugarLogger.Errorf("Failed to add %s to Admins group: %v", entityID, err)
			continue
		}
		logger.SugarLogger.Infof("Added %s to Admins group from ADMIN_ENTITY_IDS", entityID)
	}
}

// linkAdminsGroupToSentinelApp ensures the Admins group is linked to the
// Sentinel application so the Admins membership flows into every OAuth
// app's tokens (Sentinel's linked groups act as a global default for the
// claim filter). Created with required=false: this is a global default for
// the claim, not a gate on Sentinel itself.
//
// Idempotent: only creates the link if it's missing, so admins toggling
// the required flag through the UI won't be clobbered on next boot.
func linkAdminsGroupToSentinelApp() {
	groups, err := service.GetGroupsForApplication(SentinelApplicationID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to check Sentinel app group links: %v", err)
		return
	}
	for _, g := range groups {
		if g.ID == service.AdminsGroupID {
			return
		}
	}
	if _, err := service.UpsertApplicationGroup(model.ApplicationGroup{
		ApplicationID: SentinelApplicationID,
		GroupID:       service.AdminsGroupID,
		Required:      false,
	}); err != nil {
		logger.SugarLogger.Errorf("Failed to link Admins group to Sentinel app: %v", err)
		return
	}
	logger.SugarLogger.Infof("Linked Admins group to Sentinel application")
}

// SentinelCoreServiceAccountName is the kebab-case identifier this SA
// goes by. Matches the docker container name (sentinel-core) and the
// naming convention used by every other internal SA.
const SentinelCoreServiceAccountName = "sentinel-core"

func initializeDefaultServiceAccounts() {
	_, err := service.GetServiceAccountByID(SentinelCoreServiceAccountID)
	if err == gorm.ErrRecordNotFound {
		serviceAccount, err := service.CreateServiceAccount(model.ServiceAccount{
			ID:            SentinelCoreServiceAccountID,
			EntityID:      SentinelCoreEntityID,
			ApplicationID: SentinelApplicationID,
			Name:          SentinelCoreServiceAccountName,
			CreatedBy:     SentinelCoreServiceAccountID,
		})
		if err != nil {
			logger.SugarLogger.Fatalf("Failed to create Sentinel core service account: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Sentinel core service account (id=%s)", serviceAccount.ID)
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Sentinel core service account: %v", err)
	} else {
		// Migrate any existing row (older builds seeded the name as
		// "Sentinel Core" — Title Case with a space). Idempotent: the
		// UPDATE is a no-op when the value already matches.
		if err := database.DB.Model(&model.ServiceAccount{}).
			Where("id = ? AND name <> ?", SentinelCoreServiceAccountID, SentinelCoreServiceAccountName).
			Update("name", SentinelCoreServiceAccountName).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to normalize Sentinel core SA name: %v", err)
		} else {
			logger.SugarLogger.Infoln("Sentinel core service account already exists")
		}
	}
}

// initializeInternalServiceAccounts ensures each non-core service has a
// pre-seeded SA with a minted bearer token. Idempotent on both axes:
//
//   - Looks up the SA by name; only creates if missing.
//   - Mints a token only if SignedToken is empty (a fresh SA, or one
//     whose token was lost when auth_token rows were wiped).
//
// Internal SAs deliberately bypass the human-grade scope allowlist —
// service-to-service traffic needs broad access to do system work
// (group writes, token issuance, etc.). The bootstrap exchange endpoint
// is the only way to get these tokens out of core; admins can't mint
// or rotate them through the UI.
func initializeInternalServiceAccounts() {
	for _, name := range InternalServiceAccountNames {
		sa, err := service.GetServiceAccountByName(name)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sa, err = service.CreateServiceAccountForApp(
				SentinelApplicationID,
				name,
				"sentinel:all",
				0, // never expires
				SentinelCoreEntityID,
			)
			if err != nil {
				logger.SugarLogger.Errorf("Failed to create internal SA %s: %v", name, err)
				continue
			}
			logger.SugarLogger.Infof("Created internal service account %s (id=%s, entity=%s)", name, sa.ID, sa.EntityID)
		} else if err != nil {
			logger.SugarLogger.Errorf("Failed to look up internal SA %s: %v", name, err)
			continue
		}

		if sa.SignedToken == "" {
			if _, _, err := service.MintServiceAccountToken(sa); err != nil {
				logger.SugarLogger.Errorf("Failed to mint token for internal SA %s: %v", name, err)
				continue
			}
			logger.SugarLogger.Infof("Minted bootstrap token for internal service account %s", name)
		}
	}
}
