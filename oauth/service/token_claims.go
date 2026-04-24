package service

import (
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
)

type entityResponse struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	User           *idHolder `json:"user"`
	ServiceAccount *idHolder `json:"service_account"`
}

type idHolder struct {
	ID string `json:"id"`
}

type groupResponse struct {
	ID string `json:"id"`
}

// BuildTokenClaims resolves identity and group-membership claims for an entity
// by querying core. This is where future per-app / per-entity overrides will
// be applied before a token is signed.
func BuildTokenClaims(entityID string, clientID string) map[string]interface{} {
	claims := map[string]interface{}{}

	var entity entityResponse
	if err := sentinel.Get("/core/entity/"+entityID, &entity); err != nil {
		logger.SugarLogger.Errorf("Failed to load entity %s for token claims: %v", entityID, err)
		return claims
	}
	claims["entity_type"] = entity.Type
	if entity.User != nil {
		claims["user_id"] = entity.User.ID
	}
	if entity.ServiceAccount != nil {
		claims["service_account_id"] = entity.ServiceAccount.ID
	}

	groupIDs := []string{}
	var groups []groupResponse
	if err := sentinel.Get("/core/entity/"+entityID+"/groups", &groups); err == nil {
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}
	}
	claims["groups"] = groupIDs

	return claims
}
