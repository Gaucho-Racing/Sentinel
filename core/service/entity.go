package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

const SentinelServiceAccountName = "sentinel-core"

func GetEntityByID(id string) (model.Entity, error) {
	var entity model.Entity
	if err := database.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		return model.Entity{}, err
	}
	PopulateEntity(&entity)
	return entity, nil
}

func CreateEntity(entity model.Entity) (model.Entity, error) {
	if entity.ID == "" {
		entity.ID = ulid.Make().Prefixed("ent")
	}
	if err := database.DB.Create(&entity).Error; err != nil {
		return model.Entity{}, err
	}
	PopulateEntity(&entity)
	return entity, nil
}

func PopulateEntity(entity *model.Entity) {
	var err error
	entity.EmailAuth, err = GetEmailAuthForEntity(entity.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.SugarLogger.Errorf("Failed to get email auth for entity %s: %v", entity.ID, err)
	}
	entity.PhoneAuth, err = GetPhoneAuthForEntity(entity.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.SugarLogger.Errorf("Failed to get phone auth for entity %s: %v", entity.ID, err)
	}
	entity.ExternalAuths, err = GetExternalAuthForEntity(entity.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.SugarLogger.Errorf("Failed to get external auths for entity %s: %v", entity.ID, err)
	}
	if entity.Type == model.EntityTypeUser {
		user, err := GetUserByEntityID(entity.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.SugarLogger.Errorf("Failed to get user for entity %s: %v", entity.ID, err)
		}
		if user.ID != "" {
			entity.User = &user
		}
	}
	if entity.Type == model.EntityTypeServiceAccount {
		sa, err := GetServiceAccountByEntityID(entity.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.SugarLogger.Errorf("Failed to get service account for entity %s: %v", entity.ID, err)
		}
		if sa.ID != "" {
			entity.ServiceAccount = &sa
		}
	}
}

func GetEmailAuthForEntity(entityID string) (model.EntityEmail, error) {
	var auth model.EntityEmail
	if err := database.DB.Where("entity_id = ?", entityID).First(&auth).Error; err != nil {
		return model.EntityEmail{}, err
	}
	return auth, nil
}

func CreateEmailAuthForEntity(entityID string, email string, password string) (model.EntityEmail, error) {
	auth := model.EntityEmail{
		EntityID: entityID,
		Email:    email,
		Password: password,
	}
	if err := database.DB.Create(&auth).Error; err != nil {
		return model.EntityEmail{}, err
	}
	return auth, nil
}

func GetEntityByEmail(email string) (model.Entity, error) {
	var entityEmail model.EntityEmail
	if err := database.DB.Where("email = ?", email).First(&entityEmail).Error; err != nil {
		return model.Entity{}, err
	}
	entity, err := GetEntityByID(entityEmail.EntityID)
	if err != nil {
		return model.Entity{}, err
	}
	PopulateEntity(&entity)
	return entity, nil
}

func GetPhoneAuthForEntity(entityID string) (model.EntityPhone, error) {
	var auth model.EntityPhone
	if err := database.DB.Where("entity_id = ?", entityID).First(&auth).Error; err != nil {
		return model.EntityPhone{}, err
	}
	return auth, nil
}

func CreatePhoneAuthForEntity(entityID string, phoneNumber string) (model.EntityPhone, error) {
	auth := model.EntityPhone{
		EntityID:    entityID,
		PhoneNumber: phoneNumber,
	}
	if err := database.DB.Create(&auth).Error; err != nil {
		return model.EntityPhone{}, err
	}
	return auth, nil
}

func GetEntityByExternalAuth(provider string, externalID string) (model.Entity, error) {
	var auth model.EntityExternalAuth
	if err := database.DB.Where("UPPER(provider) = UPPER(?) AND external_id = ?", provider, externalID).First(&auth).Error; err != nil {
		return model.Entity{}, err
	}
	return GetEntityByID(auth.EntityID)
}

func GetExternalAuthForEntity(entityID string) ([]model.EntityExternalAuth, error) {
	auths := []model.EntityExternalAuth{}
	if err := database.DB.Where("entity_id = ?", entityID).Find(&auths).Error; err != nil {
		return []model.EntityExternalAuth{}, err
	}
	return auths, nil
}
