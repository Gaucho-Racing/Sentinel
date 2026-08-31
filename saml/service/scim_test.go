package service

import (
	"errors"
	"testing"
)

func TestValidateSCIMEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		valid    bool
	}{
		{"https://scim.us-east-1.amazonaws.com/tenant/scim/v2", true},
		{"https://scim.cn-north-1.amazonaws.com.cn/tenant/scim/v2", true},
		{"http://scim.us-east-1.amazonaws.com/tenant/scim/v2", false},
		{"https://example.com/scim/v2", false},
		{"https://scim.us-east-1.amazonaws.com:8443/scim/v2", false},
		{"https://scim.us-east-1.amazonaws.com/scim/v2?token=value", false},
	}
	for _, test := range tests {
		t.Run(test.endpoint, func(t *testing.T) {
			err := validateSCIMEndpoint(test.endpoint)
			if test.valid && err != nil {
				t.Fatal(err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidSCIMConfiguration) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestValidateProvisioningSnapshot(t *testing.T) {
	errorsFound := validateProvisioningSnapshot(ProvisioningSnapshot{Users: []ProvisioningUser{
		{EntityID: "valid", Email: "valid@example.com", FirstName: "Valid", LastName: "User"},
		{EntityID: "missing", Email: "invalid", FirstName: "", LastName: ""},
		{EntityID: "duplicate", Email: "VALID@example.com", FirstName: "Duplicate", LastName: "User"},
	}})
	if len(errorsFound) != 4 {
		t.Fatalf("errors = %#v", errorsFound)
	}
}

func TestSCIMUserMatches(t *testing.T) {
	desired := provisioningSCIMUser(ProvisioningUser{
		EntityID:  "entity",
		Email:     "user@example.com",
		FirstName: "Test",
		LastName:  "User",
	})
	remote := desired
	remote.ID = "provider-user"
	remote.UserName = "USER@example.com"
	if !scimUserMatches(remote, desired) {
		t.Fatal("equivalent user did not match")
	}
	remote.Active = false
	if scimUserMatches(remote, desired) {
		t.Fatal("inactive user matched")
	}
}
