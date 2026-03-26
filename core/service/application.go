package service

import (
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
	return applications, nil
}

func GetApplicationByID(id string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	return app, nil
}

func GetApplicationByClientID(clientID string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	return app, nil
}

func GetApplicationsByOwnerID(ownerID string) ([]model.Application, error) {
	var applications []model.Application
	if err := database.DB.Where("owner_id = ?", ownerID).Find(&applications).Error; err != nil {
		return []model.Application{}, err
	}
	return applications, nil
}

func CreateApplication(app model.Application) (model.Application, error) {
	if app.ID == "" {
		app.ID = ulid.Make().Prefixed("app")
	}
	if err := database.DB.Create(&app).Error; err != nil {
		return model.Application{}, err
	}
	return app, nil
}

func UpdateApplication(app model.Application) (model.Application, error) {
	if err := database.DB.Save(&app).Error; err != nil {
		return model.Application{}, err
	}
	return app, nil
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
