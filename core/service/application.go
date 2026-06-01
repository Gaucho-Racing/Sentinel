package service

import (
	"crypto/rand"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm/clause"
)

// AccessedApplication is an Application plus the last time the requesting
// entity signed into it. Used by GetAccessedApplicationsForEntity for the
// "recently accessed" dashboard surface.
type AccessedApplication struct {
	model.Application
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

// GetAccessedApplicationsForEntity returns the applications the entity has
// signed into, deduplicated by client_id, ordered by most-recent access.
// Server-side dedupe so users with lopsided login distributions (many logins
// for one app, few for others) still see all distinct apps. limit=0 means
// unlimited.
func GetAccessedApplicationsForEntity(entityID string, limit int) ([]AccessedApplication, error) {
	apps := []AccessedApplication{}
	sql := `
		SELECT a.*, l.last_accessed_at
		FROM application a
		INNER JOIN (
			SELECT client_id, MAX(created_at) AS last_accessed_at
			FROM entity_login
			WHERE entity_id = ?
			GROUP BY client_id
		) l ON l.client_id = a.client_id
		ORDER BY l.last_accessed_at DESC
	`
	args := []interface{}{entityID}
	if limit > 0 {
		sql += " LIMIT ?"
		args = append(args, limit)
	}
	if err := database.DB.Raw(sql, args...).Scan(&apps).Error; err != nil {
		return []AccessedApplication{}, err
	}
	for i := range apps {
		PopulateApplication(&apps[i].Application)
	}
	return apps, nil
}

func GetAllApplications() ([]model.Application, error) {
	applications := []model.Application{}
	if err := database.DB.Find(&applications).Error; err != nil {
		return []model.Application{}, err
	}
	for i := range applications {
		PopulateApplication(&applications[i])
	}
	return applications, nil
}

func GetApplicationByID(id string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func GetApplicationByClientID(clientID string) (model.Application, error) {
	var app model.Application
	if err := database.DB.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func GetApplicationsByOwnerID(ownerID string) ([]model.Application, error) {
	applications := []model.Application{}
	if err := database.DB.Where("owner_id = ?", ownerID).Find(&applications).Error; err != nil {
		return []model.Application{}, err
	}
	for i := range applications {
		PopulateApplication(&applications[i])
	}
	return applications, nil
}

func CreateApplication(app model.Application) (model.Application, error) {
	if app.ID == "" {
		app.ID = ulid.Make().Prefixed("app")
	}
	if app.ClientID == "" {
		app.ClientID = generateSecret(12)
	}
	if app.ClientSecret == "" {
		app.ClientSecret = generateSecret(64)
	}
	if err := database.DB.Create(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func UpdateApplication(app model.Application) (model.Application, error) {
	if err := database.DB.Save(&app).Error; err != nil {
		return model.Application{}, err
	}
	PopulateApplication(&app)
	return app, nil
}

func PopulateApplication(app *model.Application) {
	uris, err := GetRedirectURIsForApplication(app.ID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get redirect URIs for application %s: %v", app.ID, err)
	}
	app.RedirectURIs = uris
}

func DeleteApplication(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.Application{}).Error; err != nil {
		return err
	}
	return nil
}

func GetGroupsForApplication(applicationID string) ([]GroupWithRequired, error) {
	appGroups := []model.ApplicationGroup{}
	if err := database.DB.Where("application_id = ?", applicationID).Find(&appGroups).Error; err != nil {
		return []GroupWithRequired{}, err
	}
	groups := []GroupWithRequired{}
	for _, ag := range appGroups {
		var group model.Group
		if err := database.DB.Where("id = ?", ag.GroupID).First(&group).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to get group %s for application %s: %v", ag.GroupID, applicationID, err)
			continue
		}
		groups = append(groups, GroupWithRequired{Group: group, Required: ag.Required})
	}
	return groups, nil
}

func GetApplicationsForGroup(groupID string) ([]ApplicationWithRequired, error) {
	appGroups := []model.ApplicationGroup{}
	if err := database.DB.Where("group_id = ?", groupID).Find(&appGroups).Error; err != nil {
		return []ApplicationWithRequired{}, err
	}
	applications := []ApplicationWithRequired{}
	for _, ag := range appGroups {
		var app model.Application
		if err := database.DB.Where("id = ?", ag.ApplicationID).First(&app).Error; err != nil {
			logger.SugarLogger.Errorf("Failed to get application %s for group %s: %v", ag.ApplicationID, groupID, err)
			continue
		}
		applications = append(applications, ApplicationWithRequired{Application: app, Required: ag.Required})
	}
	return applications, nil
}

// GroupWithRequired is a Group enriched with the Required flag from its
// application_group link. Returned by GetGroupsForApplication so the link
// metadata travels with the canonical group fields and clients don't have
// to make a second fetch to resolve names.
type GroupWithRequired struct {
	model.Group
	Required bool `json:"required"`
}

// ApplicationWithRequired mirrors GroupWithRequired in the other direction.
type ApplicationWithRequired struct {
	model.Application
	Required bool `json:"required"`
}

// UpsertApplicationGroup creates the (application_id, group_id) link if it
// doesn't exist, or updates the Required flag if it does. Used in place of
// PATCH since the project convention is POST-for-both. Only `required` is
// overwritten on conflict — `created_at` is preserved as the original link
// timestamp.
func UpsertApplicationGroup(ag model.ApplicationGroup) (model.ApplicationGroup, error) {
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "application_id"}, {Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"required"}),
	}).Create(&ag).Error
	if err != nil {
		return model.ApplicationGroup{}, err
	}
	return ag, nil
}

func DeleteApplicationGroup(applicationID string, groupID string) error {
	if err := database.DB.Where("application_id = ? AND group_id = ?", applicationID, groupID).Delete(&model.ApplicationGroup{}).Error; err != nil {
		return err
	}
	return nil
}

func GetRedirectURIsForApplication(applicationID string) ([]string, error) {
	uris := []model.ApplicationRedirectURI{}
	if err := database.DB.Where("application_id = ?", applicationID).Find(&uris).Error; err != nil {
		return []string{}, err
	}
	result := make([]string, len(uris))
	for i, uri := range uris {
		result[i] = uri.RedirectURI
	}
	return result, nil
}

func CreateApplicationRedirectURI(applicationID string, redirectURI string) (model.ApplicationRedirectURI, error) {
	uri := model.ApplicationRedirectURI{
		ApplicationID: applicationID,
		RedirectURI:   redirectURI,
	}
	if err := database.DB.Create(&uri).Error; err != nil {
		return model.ApplicationRedirectURI{}, err
	}
	return uri, nil
}

func DeleteApplicationRedirectURI(applicationID string, redirectURI string) error {
	if err := database.DB.Where("application_id = ? AND redirect_uri = ?", applicationID, redirectURI).Delete(&model.ApplicationRedirectURI{}).Error; err != nil {
		return err
	}
	return nil
}

func generateSecret(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
