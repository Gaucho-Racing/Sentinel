package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gaucho-racing/sentinel/saml/model"
)

func TestResolveAssertionIdentityUsesConfiguredMappings(t *testing.T) {
	sp := ResolvedSP{ServiceProvider: model.ServiceProvider{
		Profile:      model.ProfileGeneric,
		NameIDSource: model.NameIDSourceUsername,
		NameIDFormat: model.NameIDFormatPersistent,
		AttributeMappings: model.AttributeMappings{
			{Name: "email", FriendlyName: "mail", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceEmail, OmitIfEmpty: true},
			{Name: "groups", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceGroupNames},
			{Name: "empty", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceLastName, OmitIfEmpty: true},
		},
	}}
	identity := assertionIdentity{
		EntityID:   "ent_123",
		Email:      "driver@example.com",
		Username:   "driver",
		GroupNames: []string{"aws-sandbox-readonly", "race-ops"},
	}
	preview, err := resolveAssertionIdentity(identity, sp)
	if err != nil {
		t.Fatalf("resolveAssertionIdentity returned an error: %v", err)
	}
	if preview.NameID != "driver" || preview.NameIDFormat != model.NameIDFormatPersistent {
		t.Fatalf("unexpected NameID: %#v", preview)
	}
	if len(preview.Attributes) != 2 {
		t.Fatalf("expected two attributes, got %d", len(preview.Attributes))
	}
	if !reflect.DeepEqual(preview.Attributes[1].Values, identity.GroupNames) {
		t.Fatalf("unexpected group values: %#v", preview.Attributes[1].Values)
	}
}

func TestResolveAssertionIdentityRejectsEmptyNameID(t *testing.T) {
	_, err := resolveAssertionIdentity(assertionIdentity{EntityID: "ent_123"}, ResolvedSP{
		ServiceProvider: model.ServiceProvider{
			Profile:           model.ProfileGeneric,
			NameIDSource:      model.NameIDSourceEmail,
			NameIDFormat:      model.NameIDFormatEmail,
			AttributeMappings: model.AttributeMappings{},
		},
	})
	if !errors.Is(err, ErrNameIDEmpty) {
		t.Fatalf("expected ErrNameIDEmpty, got %v", err)
	}
}

func TestResolveAssertionIdentityValidatesAWSNameIDEmail(t *testing.T) {
	_, err := resolveAssertionIdentity(assertionIdentity{Email: "Driver Name <driver@example.com>"}, ResolvedSP{
		ServiceProvider: model.ServiceProvider{
			Profile:           model.ProfileAWSIdentityCenter,
			NameIDSource:      model.NameIDSourceEmail,
			NameIDFormat:      model.NameIDFormatEmail,
			AttributeMappings: model.AttributeMappings{},
		},
	})
	if !errors.Is(err, ErrNameIDEmail) {
		t.Fatalf("expected ErrNameIDEmail, got %v", err)
	}
}
