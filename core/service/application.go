package service

import (
	"crypto/rand"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllApplications() ([]model.Application, error) {
	var applications []model.Application
	if err := database.DB.Find(&applications).Error; err != nil {
		return []model.Application{}, err
	}
	for i := range applications {
		PopulateApplication(&applications[i])
	}
	return applications, nil
}

func GetApplicationByID(id string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func GetApplicationByClientID(clientID string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func GetApplicationsByOwnerID(ownerID string) ([]model.Application, error) {
	var applications []model.Application
	if err := database.DB.Where("owner_id = ?", ownerID).Find(&applications).Error; err != nil {
		return []model.Application{}, err
	}
	for i := range applications {
		PopulateApplication(&applications[i])
	}
	return applications, nil
}

func CreateApplication(app model.Application) (model.Application, error) {
	if app.ID == "" {
		app.ID = ulid.Make().Prefixed("app")
	}
	if app.ClientID == "" {
		app.ClientID = generateSecret(12)
	}
	if app.ClientSecret == "" {
		app.ClientSecret = generateSecret(64)
	}
	if err := database.DB.Create(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func UpdateApplication(app model.Application) (model.Application, error) {
	if err := database.DB.Save(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func PopulateApplication(app *model.Application) {
	uris, err := GetRedirectURIsForApplication(app.ID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get redirect URIs for application %s: %v", app.ID, err)
	}
	app.RedirectURIs = uris
}

func DeleteApplication(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.Application{}).Error; err != nil {
		return err
	}
	return nil
}

func GetGroupsForApplication(applicationID string) ([]model.Group, error) {
	var appGroups []model.ApplicationGroup
	if err := database.DB.Where("application_id = ?", applicationID).Find(&appGroups).Error; err != nil {
		return []model.Group{}, err
	}
	var groups []model.Group
	for _, ag := range appGroups {
		var group model.Group
		if err := database.DB.Where("id = ?", ag.GroupID).First(&group).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to get group %s for application %s: %v", ag.GroupID, applicationID, err)
			continue
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func GetApplicationsForGroup(groupID string) ([]model.Application, error) {
	var appGroups []model.ApplicationGroup
	if err := database.DB.Where("group_id = ?", groupID).Find(&appGroups).Error; err != nil {
		return []model.Application{}, err
	}
	var applications []model.Application
	for _, ag := range appGroups {
		var app model.Application
		if err := database.DB.Where("id = ?", ag.ApplicationID).First(&app).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to get application %s for group %s: %v", ag.ApplicationID, groupID, err)
			continue
		}
		applications = append(applications, app)
	}
	return applications, nil
}

func CreateApplicationGroup(ag model.ApplicationGroup) (model.ApplicationGroup, error) {
	if err := database.DB.Create(&ag).Error; err != nil {
		return model.ApplicationGroup{}, err
	}
	return ag, nil
}

func DeleteApplicationGroup(applicationID string, groupID string) error {
	if err := database.DB.Where("application_id = ? AND group_id = ?", applicationID, groupID).Delete(&model.ApplicationGroup{}).Error; err != nil {
		return err
	}
	return nil
}

func GetRedirectURIsForApplication(applicationID string) ([]string, error) {
	var uris []model.ApplicationRedirectURI
	if err := database.DB.Where("application_id = ?", applicationID).Find(&uris).Error; err != nil {
		return []string{}, err
	}
	result := make([]string, len(uris))
	for i, uri := range uris {
		result[i] = uri.RedirectURI
	}
	return result, nil
}

func CreateApplicationRedirectURI(applicationID string, redirectURI string) (model.ApplicationRedirectURI, error) {
	uri := model.ApplicationRedirectURI{
		ApplicationID: applicationID,
		RedirectURI:   redirectURI,
	}
	if err := database.DB.Create(&uri).Error; err != nil {
		return model.ApplicationRedirectURI{}, err
	}
	return uri, nil
}

func DeleteApplicationRedirectURI(applicationID string, redirectURI string) error {
	if err := database.DB.Where("application_id = ? AND redirect_uri = ?", applicationID, redirectURI).Delete(&model.ApplicationRedirectURI{}).Error; err != nil {
		return err
	}
	return nil
}

func ValidateRedirectURI(applicationID string, redirectURI string) (bool, error) {
	uris, err := GetRedirectURIsForApplication(applicationID)
	if err != nil {
		return false, err
	}
	for _, uri := range uris {
		if uri == redirectURI {
			return true, nil
		}
	}
	return false, nil
}

func generateSecret(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
