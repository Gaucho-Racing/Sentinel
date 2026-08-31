package service

import (
	"database/sql"
	"testing"

	"github.com/gaucho-racing/sentinel/core/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBoundedAnalyticsValue(t *testing.T) {
	if value := boundedAnalyticsValue(0, 30, 100); value != 30 {
		t.Fatalf("default value = %d", value)
	}
	if value := boundedAnalyticsValue(101, 30, 100); value != 100 {
		t.Fatalf("maximum value = %d", value)
	}
	if value := boundedAnalyticsValue(50, 30, 100); value != 50 {
		t.Fatalf("accepted value = %d", value)
	}
}

func TestGetAnalyticsOverviewReturnsDatabaseErrors(t *testing.T) {
	sqlDB, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	if _, err := GetAnalyticsOverview(); err == nil {
		t.Fatal("expected overview query to return the database error")
	}
}
