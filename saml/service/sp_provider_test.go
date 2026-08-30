package service

import (
	"errors"
	"testing"

	"github.com/gaucho-racing/sentinel/saml/model"
)

func TestNormalizeServiceProviderDerivesMetadataFields(t *testing.T) {
	metadata := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp.example.com/metadata"><SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"><AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://sp.example.com/acs" index="0" /></SPSSODescriptor></EntityDescriptor>`
	sp, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID: "app_123",
		MetadataXML:   metadata,
	})
	if err != nil {
		t.Fatalf("NormalizeServiceProvider returned an error: %v", err)
	}
	if sp.EntityID != "https://sp.example.com/metadata" {
		t.Fatalf("unexpected entity ID: %q", sp.EntityID)
	}
	if sp.ACSURL != "https://sp.example.com/acs" {
		t.Fatalf("unexpected ACS URL: %q", sp.ACSURL)
	}
	if sp.NameIDSource != model.NameIDSourceEmail || sp.NameIDFormat != model.NameIDFormatEmail {
		t.Fatalf("unexpected default NameID mapping: %q %q", sp.NameIDSource, sp.NameIDFormat)
	}
	if len(sp.AttributeMappings) != len(DefaultAttributeMappings()) {
		t.Fatalf("expected legacy default mappings, got %d", len(sp.AttributeMappings))
	}
}

func TestNormalizeServiceProviderRejectsMetadataEntityIDMismatch(t *testing.T) {
	metadata := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp.example.com/metadata"><SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"><AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://sp.example.com/acs" index="0" /></SPSSODescriptor></EntityDescriptor>`
	_, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID: "app_123",
		EntityID:      "https://other.example.com/metadata",
		MetadataXML:   metadata,
	})
	if !errors.Is(err, ErrInvalidServiceProvider) {
		t.Fatalf("expected ErrInvalidServiceProvider, got %v", err)
	}
}

func TestNormalizeServiceProviderRejectsUnsupportedSignedRequests(t *testing.T) {
	_, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID:           "app_123",
		EntityID:                "https://sp.example.com/metadata",
		ACSURL:                  "https://sp.example.com/acs",
		WantAuthnRequestsSigned: true,
	})
	if !errors.Is(err, ErrInvalidServiceProvider) {
		t.Fatalf("expected ErrInvalidServiceProvider, got %v", err)
	}
}

func TestNormalizeServiceProviderRejectsMetadataRequiringSignedRequests(t *testing.T) {
	metadata := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp.example.com/metadata"><SPSSODescriptor AuthnRequestsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"><AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://sp.example.com/acs" index="0" /></SPSSODescriptor></EntityDescriptor>`
	_, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID: "app_123",
		MetadataXML:   metadata,
	})
	if !errors.Is(err, ErrInvalidServiceProvider) {
		t.Fatalf("expected ErrInvalidServiceProvider, got %v", err)
	}
}

func TestNormalizeServiceProviderEnforcesAWSIdentityCenterMapping(t *testing.T) {
	_, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID: "app_123",
		Profile:       model.ProfileAWSIdentityCenter,
		EntityID:      "https://aws.example.com/metadata",
		ACSURL:        "https://aws.example.com/acs",
		NameIDSource:  model.NameIDSourceEmail,
		NameIDFormat:  model.NameIDFormatEmail,
		AttributeMappings: model.AttributeMappings{
			{Name: "groups", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceGroupNames},
		},
	})
	if !errors.Is(err, ErrInvalidServiceProvider) {
		t.Fatalf("expected ErrInvalidServiceProvider, got %v", err)
	}
}

func TestNormalizeServiceProviderRejectsDuplicateAttributeNames(t *testing.T) {
	_, err := NormalizeServiceProvider(model.ServiceProvider{
		ApplicationID: "app_123",
		EntityID:      "https://sp.example.com/metadata",
		ACSURL:        "https://sp.example.com/acs",
		AttributeMappings: model.AttributeMappings{
			{Name: "email", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceEmail},
			{Name: "email", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceUsername},
		},
	})
	if !errors.Is(err, ErrInvalidServiceProvider) {
		t.Fatalf("expected ErrInvalidServiceProvider, got %v", err)
	}
}
