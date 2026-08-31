package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRevokeTokenConsumesTokenOnce(t *testing.T) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("CORE_TEST_DATABASE_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Token{}); err != nil {
		t.Fatal(err)
	}

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	token := model.Token{
		ID:        ulid.Make().Prefixed("jwt"),
		EntityID:  ulid.Make().Prefixed("ent"),
		ClientID:  "sentinel",
		Scope:     "sentinel:all refresh_token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := db.Create(&token).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", token.ID).Delete(&model.Token{}) })

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			results <- RevokeToken(token.ID)
		}()
	}
	close(start)

	var revoked int
	var alreadyConsumed int
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			revoked++
		case errors.Is(err, gorm.ErrRecordNotFound):
			alreadyConsumed++
		default:
			t.Fatalf("unexpected revoke error: %v", err)
		}
	}

	if revoked != 1 || alreadyConsumed != 1 {
		t.Fatalf("expected one successful revoke and one consumed-token error, got %d and %d", revoked, alreadyConsumed)
	}
}
