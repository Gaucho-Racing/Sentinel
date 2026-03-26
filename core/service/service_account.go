package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

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
}
