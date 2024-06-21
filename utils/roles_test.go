package utils

import (
	"sentinel/config"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	t.Run("Test Not Admin", func(t *testing.T) {
		roles := []string{"1", "2", "3"}
		if IsAdmin(roles) {
			t.Error("Expected IsAdmin to return false")
		}
	})
	t.Run("Test Is Admin", func(t *testing.T) {
		roles := []string{config.AdminRoleID, "2", "3", "4"}
		if !IsAdmin(roles) {
			t.Error("Expected IsAdmin to return true")
		}
	})
}

func TestIsOfficer(t *testing.T) {
	t.Run("Test Not Officer", func(t *testing.T) {
		roles := []string{"1", "2", "3"}
		if IsOfficer(roles) {
			t.Error("Expected IsOfficer to return false")
		}
	})
	t.Run("Test Is Officer", func(t *testing.T) {
		roles := []string{config.OfficerRoleID, "2", "3", "4"}
		if !IsOfficer(roles) {
			t.Error("Expected IsOfficer to return true")
		}
	})
}

func TestIsLead(t *testing.T) {
	t.Run("Test Not Lead", func(t *testing.T) {
		roles := []string{"1", "2", "3"}
		if IsLead(roles) {
			t.Error("Expected IsLead to return false")
		}
	})
	t.Run("Test Is Lead", func(t *testing.T) {
		roles := []string{config.LeadRoleID, "2", "3", "4"}
		if !IsLead(roles) {
			t.Error("Expected IsLead to return true")
		}
	})
}

func TestIsInnerCircle(t *testing.T) {
	t.Run("Test Not Inner Circle", func(t *testing.T) {
		roles := []string{"1", "2", "3"}
		if IsInnerCircle(roles) {
			t.Error("Expected IsInnerCircle to return false")
		}
	})
	t.Run("Test Is Inner Circle", func(t *testing.T) {
		roles := []string{config.AdminRoleID, config.OfficerRoleID, "4"}
		if !IsInnerCircle(roles) {
			t.Error("Expected IsInnerCircle to return true")
		}
	})
}
