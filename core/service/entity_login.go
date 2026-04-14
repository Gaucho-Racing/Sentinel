package service

import (
	"strconv"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
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

func GetEntityLogins(entityID string, clientID string, scope string, limit string) ([]model.EntityLogin, error) {
	var logins []model.EntityLogin
	query := database.DB.Where("entity_id = ?", entityID)
	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	query = query.Order("created_at desc")
	if limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			query = query.Limit(n)
		}
	}
	if err := query.Find(&logins).Error; err != nil {
		return []model.EntityLogin{}, err
	}
	return logins, nil
}
