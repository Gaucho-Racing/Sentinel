package service

import (
	"time"

	"github.com/gaucho-racing/sentinel/oauth/database"
	"github.com/gaucho-racing/sentinel/oauth/model"
	"github.com/gaucho-racing/ulid-go"
)

func CreateEntityLogin(login model.EntityLogin) (model.EntityLogin, error) {
	if login.ID == "" {
		login.ID = ulid.Make().Prefixed("elog")
	}
	if err := database.DB.Create(&login).Error; err != nil {
		return model.EntityLogin{}, err
	}
	return login, nil
}

func GetEntityLoginsByEntityID(entityID string) ([]model.EntityLogin, error) {
	var logins []model.EntityLogin
	if err := database.DB.Where("entity_id = ?", entityID).Order("created_at desc").Find(&logins).Error; err != nil {
		return []model.EntityLogin{}, err
	}
	return logins, nil
}

func GetLastEntityLoginForClient(entityID string, clientID string, scope string) (model.EntityLogin, error) {
	var login model.EntityLogin
	if err := database.DB.Where("entity_id = ? AND client_id = ? AND scope = ?", entityID, clientID, scope).Order("created_at desc").First(&login).Error; err != nil {
		return model.EntityLogin{}, err
	}
	return login, nil
}

func HasRecentLogin(entityID string, clientID string, scope string, within time.Duration) bool {
	login, err := GetLastEntityLoginForClient(entityID, clientID, scope)
	if err != nil {
		return false
	}
	return time.Since(login.CreatedAt) < within
}
