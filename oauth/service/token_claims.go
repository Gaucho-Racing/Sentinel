package service

import (
	"errors"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
)

// ErrAccessDenied is returned by CheckAccessGate when a user fails the
// required-group check for an application. OAuth callers should surface
// this as `access_denied` per RFC 6749.
var ErrAccessDenied = errors.New("access denied: user does not meet the required group membership for this application")

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

type applicationResponse struct {
	ID string `json:"id"`
}

type applicationGroupLink struct {
	GroupID  string `json:"group_id"`
	Required bool   `json:"required"`
}

// BuildTokenClaims resolves identity and group-membership claims for an
// entity. The groups claim is scoped by client:
//   - For the Sentinel client (config.SentinelClientID): all of the user's
//     groups are included.
//   - For any other client: the user's groups, intersected with the union
//     of (the client's linked groups) and (Sentinel's linked groups —
//     these act as a global default).
//
// Gate enforcement (CheckAccessGate) is a separate call — BuildTokenClaims
// assumes the gate has already been passed.
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

	userGroups := getEntityGroupIDs(entityID)

	if isSentinelClient(clientID) {
		claims["groups"] = userGroups
		return claims
	}

	allowed := map[string]struct{}{}
	for _, link := range getAppGroupLinks(clientID) {
		allowed[link.GroupID] = struct{}{}
	}
	for _, link := range getSentinelGroupLinks() {
		allowed[link.GroupID] = struct{}{}
	}
	filtered := make([]string, 0, len(userGroups))
	for _, g := range userGroups {
		if _, ok := allowed[g]; ok {
			filtered = append(filtered, g)
		}
	}
	claims["groups"] = filtered

	return claims
}

// CheckAccessGate returns ErrAccessDenied when the entity is not in at
// least one Required-flagged group on the application. Apps with no
// required-flagged links are open to anyone. The Sentinel client follows
// the same rule — if it has required links of its own, they apply.
func CheckAccessGate(entityID, clientID string) error {
	links := getAppGroupLinks(clientID)
	required := make([]string, 0, len(links))
	for _, link := range links {
		if link.Required {
			required = append(required, link.GroupID)
		}
	}
	if len(required) == 0 {
		return nil
	}
	userGroups := getEntityGroupIDs(entityID)
	user := make(map[string]struct{}, len(userGroups))
	for _, g := range userGroups {
		user[g] = struct{}{}
	}
	for _, g := range required {
		if _, ok := user[g]; ok {
			return nil
		}
	}
	return ErrAccessDenied
}

func isSentinelClient(clientID string) bool {
	return clientID == config.SentinelClientID
}

func getEntityGroupIDs(entityID string) []string {
	var groups []groupResponse
	if err := sentinel.Get("/core/entity/"+entityID+"/groups", &groups); err != nil {
		logger.SugarLogger.Errorf("Failed to load groups for entity %s: %v", entityID, err)
		return nil
	}
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids
}

func getAppGroupLinks(clientID string) []applicationGroupLink {
	if clientID == "" {
		return nil
	}
	var app applicationResponse
	if err := sentinel.Get("/applications/client/"+clientID, &app); err != nil {
		logger.SugarLogger.Debugf("Failed to load application for client %s: %v", clientID, err)
		return nil
	}
	var links []applicationGroupLink
	if err := sentinel.Get("/applications/"+app.ID+"/groups", &links); err != nil {
		logger.SugarLogger.Errorf("Failed to load group links for app %s: %v", app.ID, err)
		return nil
	}
	return links
}

func getSentinelGroupLinks() []applicationGroupLink {
	return getAppGroupLinks(config.SentinelClientID)
}
