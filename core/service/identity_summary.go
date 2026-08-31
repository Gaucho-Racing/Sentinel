package service

import (
	"strings"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
)

type identitySummaryRow struct {
	EntityID            string           `gorm:"column:entity_id"`
	EntityType          model.EntityType `gorm:"column:entity_type"`
	Username            string           `gorm:"column:username"`
	FirstName           string           `gorm:"column:first_name"`
	LastName            string           `gorm:"column:last_name"`
	UserAvatarURL       string           `gorm:"column:user_avatar_url"`
	ServiceAccountName  string           `gorm:"column:service_account_name"`
	ApplicationID       string           `gorm:"column:application_id"`
	ApplicationName     string           `gorm:"column:application_name"`
	ApplicationClientID string           `gorm:"column:application_client_id"`
	ApplicationIconURL  string           `gorm:"column:application_icon_url"`
}

func GetIdentitySummaries(entityIDs []string) ([]model.IdentitySummary, error) {
	entityIDs = uniqueNonEmptyStrings(entityIDs)
	if len(entityIDs) == 0 {
		return []model.IdentitySummary{}, nil
	}

	rows := []identitySummaryRow{}
	err := database.DB.
		Table("auth_entity AS entity").
		Select(`
			entity.id AS entity_id,
			entity.type AS entity_type,
			COALESCE("user".username, '') AS username,
			COALESCE("user".first_name, '') AS first_name,
			COALESCE("user".last_name, '') AS last_name,
			COALESCE("user".avatar_url, '') AS user_avatar_url,
			COALESCE(service_account.name, '') AS service_account_name,
			COALESCE(application.id, '') AS application_id,
			COALESCE(application.name, '') AS application_name,
			COALESCE(application.client_id, '') AS application_client_id,
			COALESCE(application.icon_url, '') AS application_icon_url
		`).
		Joins(`LEFT JOIN "user" ON "user".entity_id = entity.id`).
		Joins("LEFT JOIN service_account ON service_account.entity_id = entity.id").
		Joins("LEFT JOIN application ON application.id = service_account.application_id").
		Where("entity.id IN ?", entityIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return buildIdentitySummaries(entityIDs, rows), nil
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func buildIdentitySummaries(entityIDs []string, rows []identitySummaryRow) []model.IdentitySummary {
	byID := make(map[string]model.IdentitySummary, len(rows))
	for _, row := range rows {
		summary := model.IdentitySummary{
			ID:   row.EntityID,
			Type: row.EntityType,
		}
		switch row.EntityType {
		case model.EntityTypeUser:
			summary.Name = strings.TrimSpace(row.FirstName + " " + row.LastName)
			if summary.Name == "" {
				summary.Name = row.Username
			}
			summary.Username = row.Username
			summary.AvatarURL = row.UserAvatarURL
		case model.EntityTypeServiceAccount:
			summary.Name = row.ServiceAccountName
			summary.AvatarURL = row.ApplicationIconURL
			if row.ApplicationID != "" {
				summary.Application = &model.IdentityApplicationSummary{
					ID:       row.ApplicationID,
					Name:     row.ApplicationName,
					ClientID: row.ApplicationClientID,
					IconURL:  row.ApplicationIconURL,
				}
			}
		}
		if summary.Name == "" {
			summary.Name = row.EntityID
		}
		byID[row.EntityID] = summary
	}

	summaries := make([]model.IdentitySummary, 0, len(byID))
	for _, entityID := range uniqueNonEmptyStrings(entityIDs) {
		if summary, exists := byID[entityID]; exists {
			summaries = append(summaries, summary)
		}
	}
	return summaries
}
