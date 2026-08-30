package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidServiceProvider = errors.New("invalid SAML service provider configuration")

type ResolvedSP struct {
	model.ServiceProvider
	ClientID   string `json:"client_id"`
	AppName    string `json:"app_name"`
	AppIconURL string `json:"app_icon_url"`
}

type coreApplication struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
	IconURL  string `json:"icon_url"`
}

func GetServiceProvider(applicationID string) (model.ServiceProvider, error) {
	var sp model.ServiceProvider
	if err := database.DB.Where("application_id = ?", applicationID).First(&sp).Error; err != nil {
		return model.ServiceProvider{}, err
	}
	return EffectiveServiceProvider(sp), nil
}

func ResolveSP(entityID string) (ResolvedSP, error) {
	var sp model.ServiceProvider
	if err := database.DB.Where("entity_id = ?", entityID).First(&sp).Error; err != nil {
		return ResolvedSP{}, err
	}
	return resolveSP(EffectiveServiceProvider(sp))
}

func UpsertServiceProvider(sp model.ServiceProvider) (model.ServiceProvider, error) {
	sp, err := NormalizeServiceProvider(sp)
	if err != nil {
		return model.ServiceProvider{}, err
	}
	err = database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "application_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"profile", "entity_id", "acs_url", "name_id_source", "name_id_format",
			"attribute_mappings", "certificate_pem", "want_authn_requests_signed",
			"metadata_xml", "updated_at",
		}),
	}).Create(&sp).Error
	if err != nil {
		return model.ServiceProvider{}, err
	}
	return EffectiveServiceProvider(sp), nil
}

func DeleteServiceProvider(applicationID string) error {
	result := database.DB.Where("application_id = ?", applicationID).Delete(&model.ServiceProvider{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func EffectiveServiceProvider(sp model.ServiceProvider) model.ServiceProvider {
	if sp.Profile == "" {
		sp.Profile = model.ProfileGeneric
	}
	if sp.NameIDSource == "" {
		sp.NameIDSource = model.NameIDSourceEmail
	}
	if sp.NameIDFormat == "" {
		sp.NameIDFormat = model.NameIDFormatEmail
	}
	if sp.AttributeMappings == nil {
		sp.AttributeMappings = DefaultAttributeMappings()
	}
	return sp
}

func NormalizeServiceProvider(sp model.ServiceProvider) (model.ServiceProvider, error) {
	sp.ApplicationID = strings.TrimSpace(sp.ApplicationID)
	sp.EntityID = strings.TrimSpace(sp.EntityID)
	sp.ACSURL = strings.TrimSpace(sp.ACSURL)
	sp.CertificatePEM = strings.TrimSpace(sp.CertificatePEM)
	sp.MetadataXML = strings.TrimSpace(sp.MetadataXML)
	sp = EffectiveServiceProvider(sp)

	if sp.ApplicationID == "" {
		return model.ServiceProvider{}, fmt.Errorf("%w: application ID is required", ErrInvalidServiceProvider)
	}
	if sp.WantAuthnRequestsSigned {
		return model.ServiceProvider{}, fmt.Errorf("%w: signed authentication requests are not supported", ErrInvalidServiceProvider)
	}
	if sp.MetadataXML != "" {
		if sp.Profile == model.ProfileAWSIdentityCenter && len(sp.MetadataXML) > 75000 {
			return model.ServiceProvider{}, fmt.Errorf("%w: AWS IAM Identity Center metadata cannot exceed 75000 characters", ErrInvalidServiceProvider)
		}
		if len(sp.MetadataXML) > 1024*1024 {
			return model.ServiceProvider{}, fmt.Errorf("%w: metadata cannot exceed 1 MiB", ErrInvalidServiceProvider)
		}
		descriptor, err := samlsp.ParseMetadata([]byte(sp.MetadataXML))
		if err != nil {
			return model.ServiceProvider{}, fmt.Errorf("%w: parse SP metadata: %v", ErrInvalidServiceProvider, err)
		}
		if err := validateMetadataCapabilities(descriptor); err != nil {
			return model.ServiceProvider{}, err
		}
		if sp.EntityID != "" && sp.EntityID != descriptor.EntityID {
			return model.ServiceProvider{}, fmt.Errorf("%w: entity ID does not match SP metadata", ErrInvalidServiceProvider)
		}
		acsURL, err := firstHTTPPostACS(descriptor)
		if err != nil {
			return model.ServiceProvider{}, err
		}
		sp.EntityID = descriptor.EntityID
		sp.ACSURL = acsURL
	}
	if sp.EntityID == "" {
		return model.ServiceProvider{}, fmt.Errorf("%w: entity ID is required", ErrInvalidServiceProvider)
	}
	if err := validateACSURL(sp.ACSURL); err != nil {
		return model.ServiceProvider{}, err
	}
	if err := validateProfile(sp.Profile); err != nil {
		return model.ServiceProvider{}, err
	}
	if err := validateNameID(sp.NameIDSource, sp.NameIDFormat); err != nil {
		return model.ServiceProvider{}, err
	}
	if sp.Profile == model.ProfileAWSIdentityCenter && (sp.NameIDSource != model.NameIDSourceEmail || sp.NameIDFormat != model.NameIDFormatEmail) {
		return model.ServiceProvider{}, fmt.Errorf("%w: AWS IAM Identity Center requires an email NameID sourced from email", ErrInvalidServiceProvider)
	}
	if err := normalizeAttributeMappings(sp.AttributeMappings, sp.Profile); err != nil {
		return model.ServiceProvider{}, err
	}
	return sp, nil
}

func resolveSP(sp model.ServiceProvider) (ResolvedSP, error) {
	var app coreApplication
	if err := sentinel.Get("/api/applications/"+url.PathEscape(sp.ApplicationID), &app); err != nil {
		return ResolvedSP{}, fmt.Errorf("load application %s: %w", sp.ApplicationID, err)
	}
	return ResolvedSP{
		ServiceProvider: sp,
		ClientID:        app.ClientID,
		AppName:         app.Name,
		AppIconURL:      app.IconURL,
	}, nil
}

type spProvider struct{}

func (p *spProvider) GetServiceProvider(_ *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	sp, err := ResolveSP(serviceProviderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return sp.entityDescriptor()
}

func (sp ResolvedSP) entityDescriptor() (*saml.EntityDescriptor, error) {
	if sp.MetadataXML != "" {
		descriptor, err := samlsp.ParseMetadata([]byte(sp.MetadataXML))
		if err != nil {
			return nil, fmt.Errorf("parse SP metadata for %s: %w", sp.EntityID, err)
		}
		if err := validateMetadataCapabilities(descriptor); err != nil {
			return nil, err
		}
		return descriptor, nil
	}
	if sp.ACSURL == "" {
		return nil, fmt.Errorf("SP %s has neither metadata nor an ACS URL", sp.EntityID)
	}
	return &saml.EntityDescriptor{
		EntityID: sp.EntityID,
		SPSSODescriptors: []saml.SPSSODescriptor{{
			AssertionConsumerServices: []saml.IndexedEndpoint{{
				Binding:  saml.HTTPPostBinding,
				Location: sp.ACSURL,
				Index:    1,
			}},
		}},
	}, nil
}

func validateMetadataCapabilities(descriptor *saml.EntityDescriptor) error {
	for _, sp := range descriptor.SPSSODescriptors {
		if sp.AuthnRequestsSigned != nil && *sp.AuthnRequestsSigned {
			return fmt.Errorf("%w: signed authentication requests are not supported", ErrInvalidServiceProvider)
		}
	}
	return nil
}

func firstHTTPPostACS(descriptor *saml.EntityDescriptor) (string, error) {
	for _, sp := range descriptor.SPSSODescriptors {
		for _, acs := range sp.AssertionConsumerServices {
			if acs.Binding == saml.HTTPPostBinding && strings.TrimSpace(acs.Location) != "" {
				return strings.TrimSpace(acs.Location), nil
			}
		}
	}
	return "", fmt.Errorf("%w: SP metadata has no HTTP-POST assertion consumer service", ErrInvalidServiceProvider)
}

func validateACSURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return fmt.Errorf("%w: ACS URL must be an absolute HTTP or HTTPS URL", ErrInvalidServiceProvider)
	}
	return nil
}

func validateProfile(profile model.Profile) error {
	switch profile {
	case model.ProfileGeneric, model.ProfileAWSIdentityCenter:
		return nil
	default:
		return fmt.Errorf("%w: unsupported profile %q", ErrInvalidServiceProvider, profile)
	}
}

func validateNameID(source model.NameIDSource, format string) error {
	switch source {
	case model.NameIDSourceEmail, model.NameIDSourceUsername, model.NameIDSourceEntityID:
	default:
		return fmt.Errorf("%w: unsupported NameID source %q", ErrInvalidServiceProvider, source)
	}
	switch format {
	case model.NameIDFormatEmail, model.NameIDFormatPersistent, model.NameIDFormatUnspecified:
		return nil
	default:
		return fmt.Errorf("%w: unsupported NameID format %q", ErrInvalidServiceProvider, format)
	}
}

func normalizeAttributeMappings(mappings model.AttributeMappings, profile model.Profile) error {
	if len(mappings) > 50 {
		return fmt.Errorf("%w: at most 50 attribute mappings are allowed", ErrInvalidServiceProvider)
	}
	names := make(map[string]struct{}, len(mappings))
	for i := range mappings {
		mapping := &mappings[i]
		mapping.Name = strings.TrimSpace(mapping.Name)
		mapping.FriendlyName = strings.TrimSpace(mapping.FriendlyName)
		mapping.Constant = strings.TrimSpace(mapping.Constant)
		if mapping.Name == "" {
			return fmt.Errorf("%w: attribute %d has no name", ErrInvalidServiceProvider, i+1)
		}
		if _, exists := names[mapping.Name]; exists {
			return fmt.Errorf("%w: duplicate attribute name %q", ErrInvalidServiceProvider, mapping.Name)
		}
		names[mapping.Name] = struct{}{}
		if mapping.NameFormat == "" {
			mapping.NameFormat = model.AttributeNameFormatBasic
		}
		switch mapping.NameFormat {
		case model.AttributeNameFormatBasic, model.AttributeNameFormatURI, model.AttributeNameFormatUnspecified:
		default:
			return fmt.Errorf("%w: attribute %q has an unsupported name format", ErrInvalidServiceProvider, mapping.Name)
		}
		switch mapping.Source {
		case model.AttributeSourceEmail,
			model.AttributeSourceUsername,
			model.AttributeSourceFirstName,
			model.AttributeSourceLastName,
			model.AttributeSourceDisplayName,
			model.AttributeSourceEntityID,
			model.AttributeSourceGroupNames,
			model.AttributeSourceGroupIDs,
			model.AttributeSourceConstant:
		default:
			return fmt.Errorf("%w: attribute %q has an unsupported source", ErrInvalidServiceProvider, mapping.Name)
		}
		if profile == model.ProfileAWSIdentityCenter && (mapping.Source == model.AttributeSourceGroupNames || mapping.Source == model.AttributeSourceGroupIDs) {
			return fmt.Errorf("%w: AWS IAM Identity Center mappings cannot emit multi-valued group attributes", ErrInvalidServiceProvider)
		}
	}
	return nil
}
