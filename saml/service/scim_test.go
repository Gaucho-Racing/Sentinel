package service

import (
	"errors"
	"testing"
	"time"

	"github.com/gaucho-racing/sentinel/saml/model"
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

func TestSCIMSyncIntervalDuration(t *testing.T) {
	tests := map[model.SCIMSyncInterval]time.Duration{
		model.SCIMSyncInterval5Minutes:  5 * time.Minute,
		model.SCIMSyncInterval15Minutes: 15 * time.Minute,
		model.SCIMSyncInterval30Minutes: 30 * time.Minute,
		model.SCIMSyncInterval1Hour:     time.Hour,
		model.SCIMSyncInterval6Hours:    6 * time.Hour,
		model.SCIMSyncIntervalDaily:     24 * time.Hour,
	}
	for interval, expected := range tests {
		duration, err := scimSyncIntervalDuration(interval)
		if err != nil || duration != expected {
			t.Fatalf("interval %s = %v, %v", interval, duration, err)
		}
	}
	if _, err := scimSyncIntervalDuration("2h"); !errors.Is(err, ErrInvalidSCIMConfiguration) {
		t.Fatalf("invalid interval error = %v", err)
	}
}

func TestAdvanceSCIMScheduleSkipsMissedIntervals(t *testing.T) {
	scheduledAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	now := scheduledAt.Add(17 * time.Minute)
	next := advanceSCIMSchedule(scheduledAt, 5*time.Minute, now)
	expected := scheduledAt.Add(20 * time.Minute)
	if !next.Equal(expected) {
		t.Fatalf("next = %v, expected %v", next, expected)
	}
}
