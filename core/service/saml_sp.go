package service

import (
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"gorm.io/gorm/clause"
)

// ResolvedSAMLServiceProvider is a SAML SP registration joined with the
// identifying fields of its owning application. The saml service resolves an
// inbound AuthnRequest's issuer to one of these: ClientID drives the access
// gate and group filtering (the same client-scoped logic OAuth uses), and the
// app name/icon feed the consent screen.
type ResolvedSAMLServiceProvider struct {
	model.SAMLServiceProvider
	ClientID   string `json:"client_id"`
	AppName    string `json:"app_name"`
	AppIconURL string `json:"app_icon_url"`
}

func GetSAMLServiceProviderByApplicationID(applicationID string) (model.SAMLServiceProvider, error) {
	var sp model.SAMLServiceProvider
	if err := database.DB.Where("application_id = ?", applicationID).First(&sp).Error; err != nil {
		return model.SAMLServiceProvider{}, err
	}
	return sp, nil
}

func resolveSAMLServiceProvider(sp model.SAMLServiceProvider) (ResolvedSAMLServiceProvider, error) {
	app, err := GetApplicationByID(sp.ApplicationID)
	if err != nil {
		return ResolvedSAMLServiceProvider{}, err
	}
	return ResolvedSAMLServiceProvider{
		SAMLServiceProvider: sp,
		ClientID:            app.ClientID,
		AppName:             app.Name,
		AppIconURL:          app.IconURL,
	}, nil
}

func GetResolvedSAMLServiceProviderByEntityID(entityID string) (ResolvedSAMLServiceProvider, error) {
	var sp model.SAMLServiceProvider
	if err := database.DB.Where("entity_id = ?", entityID).First(&sp).Error; err != nil {
		return ResolvedSAMLServiceProvider{}, err
	}
	return resolveSAMLServiceProvider(sp)
}

// UpsertSAMLServiceProvider creates the SP registration for an application or
// updates it in place. Keyed on application_id so an app has at most one SAML
// SP config; created_at is preserved on conflict.
func UpsertSAMLServiceProvider(sp model.SAMLServiceProvider) (model.SAMLServiceProvider, error) {
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "application_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"entity_id", "acs_url", "name_id_format",
			"certificate_pem", "want_authn_requests_signed", "metadata_xml", "updated_at",
		}),
	}).Create(&sp).Error
	if err != nil {
		return model.SAMLServiceProvider{}, err
	}
	return sp, nil
}

func DeleteSAMLServiceProvider(applicationID string) error {
	return database.DB.Where("application_id = ?", applicationID).Delete(&model.SAMLServiceProvider{}).Error
}
