package service

import (
	"os"
	"testing"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetApplicationProvisioningSnapshot(t *testing.T) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("CORE_TEST_DATABASE_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.Entity{},
		&model.EntityEmail{},
		&model.User{},
		&model.Group{},
		&model.GroupMember{},
		&model.ApplicationGroup{},
	); err != nil {
		t.Fatal(err)
	}
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	now := time.Now().UTC()
	records := []any{
		&model.Entity{ID: "user-active", Type: model.EntityTypeUser},
		&model.Entity{ID: "user-expired", Type: model.EntityTypeUser},
		&model.Entity{ID: "service", Type: model.EntityTypeServiceAccount},
		&model.EntityEmail{EntityID: "user-active", Email: "active@example.com"},
		&model.EntityEmail{EntityID: "user-expired", Email: "expired@example.com"},
		&model.User{ID: "profile-active", EntityID: "user-active", Username: "active", FirstName: "Active", LastName: "User"},
		&model.User{ID: "profile-expired", EntityID: "user-expired", Username: "expired", FirstName: "Expired", LastName: "User"},
		&model.Group{ID: "group-one", Name: "Group One"},
		&model.Group{ID: "group-two", Name: "Group Two"},
		&model.Group{ID: "group-unlinked", Name: "Unlinked"},
		&model.ApplicationGroup{ApplicationID: "application", GroupID: "group-one"},
		&model.ApplicationGroup{ApplicationID: "application", GroupID: "group-two"},
		&model.GroupMember{GroupID: "group-one", EntityID: "user-active"},
		&model.GroupMember{GroupID: "group-one", EntityID: "user-expired", HasExpiration: true, ExpiresAt: now.Add(-time.Hour)},
		&model.GroupMember{GroupID: "group-one", EntityID: "service"},
		&model.GroupMember{GroupID: "group-two", EntityID: "user-active"},
	}
	for _, record := range records {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}

	snapshot, err := GetApplicationProvisioningSnapshot("application")
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Users) != 1 || snapshot.Users[0].EntityID != "user-active" {
		t.Fatalf("users = %#v", snapshot.Users)
	}
	if len(snapshot.Groups) != 2 {
		t.Fatalf("groups = %#v", snapshot.Groups)
	}
	for _, group := range snapshot.Groups {
		if len(group.Members) != 1 || group.Members[0] != "user-active" {
			t.Fatalf("group = %#v", group)
		}
	}
}
