package jobs

import (
	"strings"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"gorm.io/gorm"
)

const SentinelClientID = "sentinel"
const SentinelApplicationID = "app_01kpy5f8263c4rqnhn9v2akdvf"

const SentinelCoreEntityID = "ent_01kpy5f8263c4rqnhn9y920fkn"
const SentinelCoreServiceAccountID = "sa_01kpy5f8263c4rqnhn9zejxyhk"

func InitializeCore() {
	initializeDefaultApplications()
	initializeDefaultEntities()
	initializeDefaultServiceAccounts()
	initializeAdminsGroup()
	linkAdminsGroupToSentinelApp()
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

func initializeDefaultServiceAccounts() {
	_, err := service.GetServiceAccountByID(SentinelCoreServiceAccountID)
	if err == gorm.ErrRecordNotFound {
		serviceAccount, err := service.CreateServiceAccount(model.ServiceAccount{
			ID:            SentinelCoreServiceAccountID,
			EntityID:      SentinelCoreEntityID,
			ApplicationID: SentinelApplicationID,
			Name:          "Sentinel Core",
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
		logger.SugarLogger.Infoln("Sentinel core service account already exists")
	}
}
