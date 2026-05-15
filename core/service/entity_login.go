package service

import (
	"strconv"
	"time"

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

// EntityLoginsFilter holds the query params accepted by GetEntityLogins.
// All string fields are optional; empty strings are ignored.
type EntityLoginsFilter struct {
	EntityID string
	ClientID string
	Scope    string
	Before   string // RFC3339; matches logins with created_at < Before
	After    string // RFC3339; matches logins with created_at > After
	Limit    string // integer string; 0 or unset means unlimited
}

func GetEntityLogins(filter EntityLoginsFilter) ([]model.EntityLogin, error) {
	logins := []model.EntityLogin{}
	query := database.DB.Where("entity_id = ?", filter.EntityID)
	if filter.ClientID != "" {
		query = query.Where("client_id = ?", filter.ClientID)
	}
	if filter.Scope != "" {
		query = query.Where("scope = ?", filter.Scope)
	}
	if filter.Before != "" {
		if t, err := time.Parse(time.RFC3339, filter.Before); err == nil {
			query = query.Where("created_at < ?", t)
		}
	}
	if filter.After != "" {
		if t, err := time.Parse(time.RFC3339, filter.After); err == nil {
			query = query.Where("created_at > ?", t)
		}
	}
	query = query.Order("created_at desc")
	if filter.Limit != "" {
		if n, err := strconv.Atoi(filter.Limit); err == nil && n > 0 {
			query = query.Limit(n)
		}
	}
	if err := query.Find(&logins).Error; err != nil {
		return []model.EntityLogin{}, err
	}
	return logins, nil
}
