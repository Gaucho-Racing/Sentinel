package jobs

import (
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"gorm.io/gorm"
)

const SentinelClientID = "sentinel"

const SentinelCoreEntityID = "ent_01kpgkjbstpswced3c61rjrbkh"
const SentinelCoreServiceAccountID = "sa_01kpgkhs9k6mxkaqff0tmtqm0y"

func InitializeCore() {
	initializeDefaultApplications()
	initializeDefaultEntities()
	// initializeDefaultServiceAccounts()
}

func initializeDefaultApplications() {
	_, err := service.GetApplicationByClientID(SentinelClientID)
	if err == gorm.ErrRecordNotFound {
		app, err := service.CreateApplication(model.Application{
			Name:        "Sentinel",
			Description: "Gaucho Racing's authentication service",
			ClientID:    SentinelClientID,
		})
		if err != nil {
			logger.SugarLogger.Fatalf("Failed to create Sentinel application: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Sentinel application (id=%s, client_id=%s)", app.ID, app.ClientID)
		logger.SugarLogger.Infof("Sentinel client secret: %s", app.ClientSecret)

		defaultRedirectURIs := []string{
			"http://localhost:3000/auth/callback",
			"https://sso.gauchoracing.com/auth/callback",
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

// sa_01kpgkhs9k6mxkaqff0tmtqm0y
// ent_01kpgkjbstpswced3c61rjrbkh
func initializeDefaultEntities() {
	sentinelCoreEntityID := "ent_01kpgkjbstpswced3c61rjrbkh"
	sentinelCoreServiceAccountID := "sa_01kpgkhs9k6mxkaqff0tmtqm0y"
	var sentinelCoreEntity model.Entity
	var sentinelCoreServiceAccount model.ServiceAccount
	var err error
	sentinelCoreEntity, err = service.GetEntityByID(sentinelCoreEntityID)
	if err == gorm.ErrRecordNotFound {
		sentinelCoreEntity, err := service.CreateEntity(model.Entity{
			ID:   sentinelCoreEntity.ID,
			Type: model.EntityTypeServiceAccount,
		})
		if err != nil {
			logger.SugarLogger.Fatalf("Failed to create Sentinel core entity: %v", err)
			return
		}
		logger.SugarLogger.Infof("Created Sentinel core entity (id=%s)", sentinelCoreEntity.ID)
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Sentinel core entity: %v", err)
	} else {
		logger.SugarLogger.Infoln("Sentinel core entity already exists")
	}
	sentinelCoreServiceAccount, err = service.GetServiceAccountByID(sentinelCoreServiceAccountID)
	if err == gorm.ErrRecordNotFound {
		sentinelCoreServiceAccount, err = service.CreateServiceAccount(model.ServiceAccount{
			ID:       sentinelCoreServiceAccountID,
			EntityID: sentinelCoreEntity.ID,
			Name:     "Sentinel Core",
		})
	}
	print(sentinelCoreServiceAccount.ID)
}
