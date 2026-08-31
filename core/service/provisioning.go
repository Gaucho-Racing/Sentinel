package service

import (
	"sort"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
)

type ProvisioningUser struct {
	EntityID  string `json:"entity_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ProvisioningGroup struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type ProvisioningSnapshot struct {
	Users  []ProvisioningUser  `json:"users"`
	Groups []ProvisioningGroup `json:"groups"`
}

type provisioningMemberRow struct {
	GroupID   string
	EntityID  string
	Email     string
	Username  string
	FirstName string
	LastName  string
}

func GetApplicationProvisioningSnapshot(applicationID string) (ProvisioningSnapshot, error) {
	groups := []ProvisioningGroup{}
	if err := database.DB.Raw(`
		SELECT g.id, g.name
		FROM application_group ag
		JOIN "group" g ON g.id = ag.group_id
		WHERE ag.application_id = ?
		ORDER BY g.name, g.id
	`, applicationID).Scan(&groups).Error; err != nil {
		return ProvisioningSnapshot{}, err
	}

	if len(groups) == 0 {
		return ProvisioningSnapshot{Users: []ProvisioningUser{}, Groups: groups}, nil
	}

	groupIDs := make([]string, len(groups))
	for index := range groups {
		groupIDs[index] = groups[index].ID
		groups[index].Members = []string{}
	}

	rows := []provisioningMemberRow{}
	if err := database.DB.Raw(`
		SELECT gm.group_id, gm.entity_id, COALESCE(ee.email, '') AS email,
		       COALESCE(u.username, '') AS username, COALESCE(u.first_name, '') AS first_name,
		       COALESCE(u.last_name, '') AS last_name
		FROM group_member gm
		JOIN auth_entity e ON e.id = gm.entity_id AND e.type = 'USER'
		LEFT JOIN "user" u ON u.entity_id = gm.entity_id
		LEFT JOIN auth_entity_email ee ON ee.entity_id = gm.entity_id
		WHERE gm.group_id IN ?
		  AND (gm.has_expiration = false OR gm.expires_at > ?)
		ORDER BY gm.group_id, gm.entity_id
	`, groupIDs, time.Now().UTC()).Scan(&rows).Error; err != nil {
		return ProvisioningSnapshot{}, err
	}

	groupIndexes := make(map[string]int, len(groups))
	for index := range groups {
		groupIndexes[groups[index].ID] = index
	}
	usersByID := make(map[string]ProvisioningUser)
	for _, row := range rows {
		groupIndex, ok := groupIndexes[row.GroupID]
		if !ok {
			continue
		}
		groups[groupIndex].Members = append(groups[groupIndex].Members, row.EntityID)
		usersByID[row.EntityID] = ProvisioningUser{
			EntityID:  row.EntityID,
			Email:     row.Email,
			Username:  row.Username,
			FirstName: row.FirstName,
			LastName:  row.LastName,
		}
	}

	users := make([]ProvisioningUser, 0, len(usersByID))
	for _, user := range usersByID {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].EntityID < users[j].EntityID })

	return ProvisioningSnapshot{Users: users, Groups: groups}, nil
}
