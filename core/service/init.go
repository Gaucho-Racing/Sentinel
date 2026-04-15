package service

import (
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/gorm"
)

const SentinelClientID = "sentinel"

func InitializeDefaultApplications() {
	_, err := GetApplicationByClientID(SentinelClientID)
	if err == gorm.ErrRecordNotFound {
		app, err := CreateApplication(model.Application{
			Name:        "Sentinel",
			Description: "Sentinel Identity Platform",
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
			CreateApplicationRedirectURI(app.ID, uri)
		}
		logger.SugarLogger.Infof("Added %d default redirect URIs", len(defaultRedirectURIs))
	} else if err != nil {
		logger.SugarLogger.Fatalf("Failed to check for Sentinel application: %v", err)
	} else {
		logger.SugarLogger.Infoln("Sentinel application already exists")
	}
}
