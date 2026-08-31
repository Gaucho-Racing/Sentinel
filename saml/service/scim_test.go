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
	snapshot, skippedUsers := prepareProvisioningSnapshot(ProvisioningSnapshot{
		Users: []ProvisioningUser{
			{EntityID: "valid", Email: "valid@example.com", FirstName: "Valid", LastName: "User"},
			{EntityID: "missing", Email: "invalid", Username: "owen", FirstName: "", LastName: ""},
			{EntityID: "duplicate", Email: "VALID@example.com", FirstName: "Duplicate", LastName: "User"},
		},
		Groups: []ProvisioningGroup{{ID: "group", Name: "AWS Admins", Members: []string{"valid", "missing"}}},
	})
	if len(skippedUsers) != 1 || skippedUsers[0].EntityID != "missing" || skippedUsers[0].Username != "owen" {
		t.Fatalf("skipped users = %#v", skippedUsers)
	}
	if len(skippedUsers[0].Groups) != 1 || skippedUsers[0].Groups[0] != "AWS Admins" {
		t.Fatalf("skipped user groups = %#v", skippedUsers[0].Groups)
	}
	if len(snapshot.Users) != 2 {
		t.Fatalf("users = %#v", snapshot.Users)
	}
	if len(snapshot.Groups[0].Members) != 1 || snapshot.Groups[0].Members[0] != "valid" {
		t.Fatalf("members = %#v", snapshot.Groups[0].Members)
	}
	errorsFound := validateProvisioningSnapshot(snapshot)
	if len(errorsFound) != 1 {
		t.Fatalf("errors = %#v", errorsFound)
	}
}

func TestPrepareProvisioningSnapshotRejectsDisplayNameEmail(t *testing.T) {
	snapshot, skippedUsers := prepareProvisioningSnapshot(ProvisioningSnapshot{Users: []ProvisioningUser{
		{EntityID: "display-name", Email: "Owen <owen@example.com>", FirstName: "Owen", LastName: "User"},
	}})
	if len(snapshot.Users) != 0 || len(skippedUsers) != 1 {
		t.Fatalf("snapshot = %#v, skipped users = %#v", snapshot, skippedUsers)
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
