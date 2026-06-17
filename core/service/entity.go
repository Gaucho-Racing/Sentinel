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

// DeleteEntity removes the bare Entity row. Cascade of associated rows
// (email_auth, phone_auth, external_auth, user, service_account, group
// memberships) is intentionally NOT done here — callers that delete a
// derived row (e.g. ServiceAccount) decide whether to also delete the
// underlying entity, so the rules are explicit at the call site.
func DeleteEntity(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.Entity{}).Error
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

func UpdateEmailAuthForEntity(entityID string, email string, password string) (model.EntityEmail, error) {
	var auth model.EntityEmail
	if err := database.DB.Where("entity_id = ?", entityID).First(&auth).Error; err != nil {
		return model.EntityEmail{}, err
	}
	auth.Email = email
	auth.Password = password
	if err := database.DB.Save(&auth).Error; err != nil {
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

// ListExternalAuthsByProvider returns every external auth row for the given
// provider (e.g. "DISCORD"). Used by integration services to enumerate
// onboarded users for a provider without having to walk all entities.
func ListExternalAuthsByProvider(provider string) ([]model.EntityExternalAuth, error) {
	auths := []model.EntityExternalAuth{}
	if err := database.DB.Where("UPPER(provider) = UPPER(?)", provider).Find(&auths).Error; err != nil {
		return []model.EntityExternalAuth{}, err
	}
	return auths, nil
}

func CreateExternalAuthForEntity(auth model.EntityExternalAuth) (model.EntityExternalAuth, error) {
	if err := database.DB.Create(&auth).Error; err != nil {
		return model.EntityExternalAuth{}, err
	}
	return auth, nil
}

// UpdateExternalAuthMetadata refreshes the provider-supplied metadata jsonb on
// an existing external auth row. Used by login handlers to keep email /
// username / avatar / etc. current across sessions. No-op (returns
// gorm.ErrRecordNotFound) if no row matches — callers decide whether that's
// fatal.
func UpdateExternalAuthMetadata(entityID string, provider string, metadata model.JSONMap) error {
	result := database.DB.Model(&model.EntityExternalAuth{}).
		Where("entity_id = ? AND UPPER(provider) = UPPER(?)", entityID, provider).
		Update("metadata", metadata)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
