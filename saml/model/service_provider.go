package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Profile string

const (
	ProfileGeneric           Profile = "GENERIC"
	ProfileAWSIdentityCenter Profile = "AWS_IDENTITY_CENTER"
)

type NameIDSource string

const (
	NameIDSourceEmail              NameIDSource = "EMAIL"
	NameIDSourceUsername           NameIDSource = "USERNAME"
	NameIDSourceEntityID           NameIDSource = "ENTITY_ID"
	NameIDFormatEmail                           = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	NameIDFormatPersistent                      = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	NameIDFormatUnspecified                     = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
	AttributeNameFormatBasic                    = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
	AttributeNameFormatURI                      = "urn:oasis:names:tc:SAML:2.0:attrname-format:uri"
	AttributeNameFormatUnspecified              = "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified"
)

type AttributeSource string

const (
	AttributeSourceEmail       AttributeSource = "EMAIL"
	AttributeSourceUsername    AttributeSource = "USERNAME"
	AttributeSourceFirstName   AttributeSource = "FIRST_NAME"
	AttributeSourceLastName    AttributeSource = "LAST_NAME"
	AttributeSourceDisplayName AttributeSource = "DISPLAY_NAME"
	AttributeSourceEntityID    AttributeSource = "ENTITY_ID"
	AttributeSourceGroupNames  AttributeSource = "GROUP_NAMES"
	AttributeSourceGroupIDs    AttributeSource = "GROUP_IDS"
	AttributeSourceConstant    AttributeSource = "CONSTANT"
)

type AttributeMapping struct {
	Name         string          `json:"name"`
	FriendlyName string          `json:"friendly_name"`
	NameFormat   string          `json:"name_format"`
	Source       AttributeSource `json:"source"`
	Constant     string          `json:"constant"`
	OmitIfEmpty  bool            `json:"omit_if_empty"`
}

type AttributeMappings []AttributeMapping

func (m AttributeMappings) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *AttributeMappings) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

type ServiceProvider struct {
	ApplicationID           string            `json:"application_id" gorm:"primaryKey"`
	Profile                 Profile           `json:"profile"`
	EntityID                string            `json:"entity_id" gorm:"uniqueIndex"`
	ACSURL                  string            `json:"acs_url"`
	NameIDSource            NameIDSource      `json:"name_id_source"`
	NameIDFormat            string            `json:"name_id_format"`
	AttributeMappings       AttributeMappings `json:"attribute_mappings" gorm:"type:jsonb"`
	CertificatePEM          string            `json:"certificate_pem"`
	WantAuthnRequestsSigned bool              `json:"want_authn_requests_signed"`
	MetadataXML             string            `json:"metadata_xml"`
	UpdatedAt               time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt               time.Time         `json:"created_at" gorm:"autoCreateTime"`
}

func (ServiceProvider) TableName() string {
	return "saml_service_provider"
}
