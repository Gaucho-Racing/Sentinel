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

func TestReviewJoinRequestScopesParentAndCommitsMembershipAtomically(t *testing.T) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("CORE_TEST_DATABASE_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.GroupJoinRequest{}, &model.GroupJoinRequestComment{}, &model.GroupMember{}); err != nil {
		t.Fatal(err)
	}

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	requestID := ulid.Make().Prefixed("gjr")
	groupID := ulid.Make().Prefixed("grp")
	wrongGroupID := ulid.Make().Prefixed("grp")
	entityID := ulid.Make().Prefixed("ent")
	request := model.GroupJoinRequest{
		ID:       requestID,
		GroupID:  groupID,
		EntityID: entityID,
		Status:   string(model.GroupJoinRequestStatusPending),
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("group_id = ? AND entity_id = ?", groupID, entityID).Delete(&model.GroupMember{})
		db.Where("id = ?", requestID).Delete(&model.GroupJoinRequest{})
	})

	if _, err := ReviewJoinRequest(wrongGroupID, requestID, "reviewer", model.GroupJoinRequestStatusApproved, false, time.Time{}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected parent mismatch to return not found, got %v", err)
	}

	var unchanged model.GroupJoinRequest
	if err := db.Where("id = ?", requestID).First(&unchanged).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.Status != string(model.GroupJoinRequestStatusPending) {
		t.Fatalf("request changed after mismatched parent: %s", unchanged.Status)
	}

	reviewed, err := ReviewJoinRequest(groupID, requestID, "reviewer", model.GroupJoinRequestStatusApproved, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Status != string(model.GroupJoinRequestStatusApproved) || reviewed.ReviewedBy != "reviewer" {
		t.Fatalf("unexpected reviewed request: %#v", reviewed)
	}

	var member model.GroupMember
	if err := db.Where("group_id = ? AND entity_id = ?", groupID, entityID).First(&member).Error; err != nil {
		t.Fatal(err)
	}
	if member.AddedBy != "reviewer" {
		t.Fatalf("membership actor = %q", member.AddedBy)
	}

	if _, err := ReviewJoinRequest(groupID, requestID, "other", model.GroupJoinRequestStatusApproved, false, time.Time{}); !errors.Is(err, ErrJoinRequestNotPending) {
		t.Fatalf("expected repeated review to be rejected, got %v", err)
	}
}
