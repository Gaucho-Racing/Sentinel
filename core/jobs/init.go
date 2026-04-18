package jobs

import (
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"gorm.io/gorm"
)

func InitializeCore() {
	initializeDefaultApplications()
	initializeDefaultServiceAccount()
}

func initializeDefaultApplications() {
	_, err := service.GetApplicationByClientID(service.SentinelClientID)
	if err == gorm.ErrRecordNotFound {
		app, err := service.CreateApplication(model.Application{
			Name:        "Sentinel",
			Description: "Sentinel Identity Platform",
			ClientID:    service.SentinelClientID,
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

func initializeDefaultServiceAccount() {
	app, err := service.GetApplicationByClientID(service.SentinelClientID)
	if err != nil {
		logger.SugarLogger.Fatalf("Sentinel application not found, cannot create service account: %v", err)
		return
	}

	existing, _ := service.GetServiceAccountByName(service.SentinelServiceAccountName)
	if existing.ID != "" {
		logger.SugarLogger.Infoln("Sentinel service account already exists")
		return
	}

	entity, err := service.CreateEntity(model.Entity{
		Type: model.EntityTypeServiceAccount,
	})
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to create entity for Sentinel service account: %v", err)
		return
	}

	sa, err := service.CreateServiceAccount(model.ServiceAccount{
		EntityID:      entity.ID,
		ApplicationID: app.ID,
		Name:          service.SentinelServiceAccountName,
		CreatedBy:     "system",
	})
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to create Sentinel service account: %v", err)
		return
	}
	logger.SugarLogger.Infof("Created Sentinel service account (id=%s, entity_id=%s)", sa.ID, sa.EntityID)

	token, tokenID, err := service.GenerateToken(entity.ID, app.ClientID, "sentinel:all", 365*24*3600, nil)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to generate token for Sentinel service account: %v", err)
		return
	}
	logger.SugarLogger.Infof("Generated Sentinel service account token (id=%s)", tokenID)
	logger.SugarLogger.Infof("Sentinel service account token: %s", token)
}
