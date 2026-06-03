package model

import "time"

// SAMLServiceProvider extends an Application with the registration data a SAML
// relying party needs. The owning Application row still carries the
// name/icon/owner and the group links that drive the access gate and group
// filtering — exactly as it does for OAuth clients — so this table only holds
// what SAML adds on top.
//
// MetadataXML, when present, is the SP's published metadata and is the source
// of truth for the ACS endpoint and signing certificate. The discrete fields
// (EntityID, ACSURL, CertificatePEM) are kept populated for SPs registered
// manually without metadata, and EntityID is always set so the SSO endpoint
// can resolve an inbound AuthnRequest's issuer back to its application.
type SAMLServiceProvider struct {
	ApplicationID           string    `json:"application_id" gorm:"primaryKey"`
	EntityID                string    `json:"entity_id" gorm:"uniqueIndex"`
	ACSURL                  string    `json:"acs_url"`
	NameIDFormat            string    `json:"name_id_format"`
	CertificatePEM          string    `json:"certificate_pem"`
	WantAuthnRequestsSigned bool      `json:"want_authn_requests_signed"`
	MetadataXML             string    `json:"metadata_xml"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SAMLServiceProvider) TableName() string {
	return "saml_service_provider"
}
