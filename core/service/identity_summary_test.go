package service

import (
	"os"
	"reflect"
	"testing"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetIdentitySummaries(t *testing.T) {
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
		&model.User{},
		&model.Application{},
		&model.ServiceAccount{},
	); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	originalDB := database.DB
	database.DB = tx
	defer func() {
		database.DB = originalDB
		tx.Rollback()
	}()

	records := []any{
		&model.Entity{ID: "ent_identity_summary_user", Type: model.EntityTypeUser},
		&model.Entity{ID: "ent_identity_summary_service", Type: model.EntityTypeServiceAccount},
		&model.User{
			ID:        "usr_identity_summary",
			EntityID:  "ent_identity_summary_user",
			Username:  "summary-user",
			FirstName: "Summary",
			LastName:  "User",
			AvatarURL: "https://example.com/user.png",
		},
		&model.Application{
			ID:       "app_identity_summary",
			Name:     "Summary Application",
			ClientID: "summary-client",
			IconURL:  "https://example.com/application.png",
		},
		&model.ServiceAccount{
			ID:            "sa_identity_summary",
			EntityID:      "ent_identity_summary_service",
			ApplicationID: "app_identity_summary",
			Name:          "summary-worker",
		},
	}
	for _, record := range records {
		if err := tx.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}

	summaries, err := GetIdentitySummaries([]string{
		"ent_identity_summary_service",
		"ent_identity_summary_user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries = %#v", summaries)
	}
	if summaries[0].Name != "summary-worker" || summaries[0].Application == nil {
		t.Fatalf("service account summary = %#v", summaries[0])
	}
	if summaries[1].Name != "Summary User" || summaries[1].AvatarURL != "https://example.com/user.png" {
		t.Fatalf("user summary = %#v", summaries[1])
	}
}

func TestBuildIdentitySummaries(t *testing.T) {
	rows := []identitySummaryRow{
		{
			EntityID:      "ent_user",
			EntityType:    model.EntityTypeUser,
			Username:      "driver",
			FirstName:     "Alex",
			LastName:      "Rivera",
			UserAvatarURL: "https://example.com/alex.png",
		},
		{
			EntityID:            "ent_service",
			EntityType:          model.EntityTypeServiceAccount,
			ServiceAccountName:  "telemetry-worker",
			ApplicationID:       "app_telemetry",
			ApplicationName:     "Telemetry",
			ApplicationClientID: "telemetry-client",
			ApplicationIconURL:  "https://example.com/telemetry.png",
		},
	}

	summaries := buildIdentitySummaries(
		[]string{"ent_service", "ent_missing", "ent_user", "ent_service"},
		rows,
	)

	if len(summaries) != 2 {
		t.Fatalf("summaries = %#v", summaries)
	}
	if summaries[0].ID != "ent_service" || summaries[1].ID != "ent_user" {
		t.Fatalf("summary order = %#v", summaries)
	}
	if summaries[0].Name != "telemetry-worker" || summaries[0].AvatarURL != "https://example.com/telemetry.png" {
		t.Fatalf("service account summary = %#v", summaries[0])
	}
	wantApplication := &model.IdentityApplicationSummary{
		ID:       "app_telemetry",
		Name:     "Telemetry",
		ClientID: "telemetry-client",
		IconURL:  "https://example.com/telemetry.png",
	}
	if !reflect.DeepEqual(summaries[0].Application, wantApplication) {
		t.Fatalf("application = %#v", summaries[0].Application)
	}
	if summaries[1].Name != "Alex Rivera" || summaries[1].Username != "driver" {
		t.Fatalf("user summary = %#v", summaries[1])
	}
}

func TestBuildIdentitySummariesFallsBackToStableID(t *testing.T) {
	summaries := buildIdentitySummaries(
		[]string{"ent_profileless"},
		[]identitySummaryRow{{EntityID: "ent_profileless", EntityType: model.EntityTypeUser}},
	)

	if len(summaries) != 1 || summaries[0].Name != "ent_profileless" {
		t.Fatalf("summaries = %#v", summaries)
	}
}
