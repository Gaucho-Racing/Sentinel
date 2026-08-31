package service

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSCIMSyncQueuePostgres(t *testing.T) {
	dsn := os.Getenv("SCIM_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SCIM_TEST_DATABASE_DSN is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.ServiceProvider{}, &model.SCIMConfiguration{}, &model.SCIMSyncRun{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_scim_sync_run_one_active_test
		ON saml_scim_sync_run (application_id)
		WHERE status IN ('QUEUED', 'RUNNING')`).Error; err != nil {
		t.Fatal(err)
	}

	previousDB := database.DB
	database.DB = db
	defer func() { database.DB = previousDB }()

	applicationID := ulid.Make().Prefixed("app")
	defer db.Where("application_id = ?", applicationID).Delete(&model.SCIMSyncRun{})
	defer db.Where("application_id = ?", applicationID).Delete(&model.SCIMConfiguration{})
	defer db.Where("application_id = ?", applicationID).Delete(&model.ServiceProvider{})

	if err := db.Create(&model.ServiceProvider{ApplicationID: applicationID, Profile: model.ProfileAWSIdentityCenter, EntityID: "urn:test:" + applicationID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SCIMConfiguration{
		ApplicationID:  applicationID,
		Endpoint:       "https://scim.us-west-2.amazonaws.com/test/scim/v2",
		AccessToken:    "secret",
		Enabled:        true,
		LastSyncStatus: model.SCIMSyncStatusNever,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := initializeSCIMSchedules(); err != nil {
		t.Fatal(err)
	}
	var initializedConfiguration model.SCIMConfiguration
	if err := db.Where("application_id = ?", applicationID).First(&initializedConfiguration).Error; err != nil {
		t.Fatal(err)
	}
	if initializedConfiguration.SyncInterval != model.SCIMSyncInterval1Hour || initializedConfiguration.NextSyncAt == nil {
		t.Fatalf("initialized configuration = %#v", initializedConfiguration)
	}

	first, created, err := EnqueueSCIMSync(context.Background(), applicationID, model.SCIMSyncTriggerManual)
	if err != nil {
		t.Fatal(err)
	}
	if !created || first.Status != model.SCIMSyncRunStatusQueued {
		t.Fatalf("first run = %#v, created = %t", first, created)
	}
	second, created, err := EnqueueSCIMSync(context.Background(), applicationID, model.SCIMSyncTriggerManual)
	if err != nil {
		t.Fatal(err)
	}
	if created || second.ID != first.ID {
		t.Fatalf("second run = %#v, created = %t", second, created)
	}

	claimed, found, err := claimSCIMSyncRun(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if !found || claimed.ID != first.ID || claimed.Status != model.SCIMSyncRunStatusRunning {
		t.Fatalf("claimed run = %#v, found = %t", claimed, found)
	}
	if err := finishSCIMSyncRun(claimed, SCIMSyncResult{DesiredUsers: 3, DesiredGroups: 2}, nil); err != nil {
		t.Fatal(err)
	}

	runs, err := ListSCIMSyncRuns(applicationID, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].Status != model.SCIMSyncRunStatusSucceeded || runs[0].Result.DesiredUsers != 3 {
		t.Fatalf("runs = %#v", runs)
	}
	if runs[0].Result.SkippedUsers == nil || runs[0].Result.ValidationErrors == nil {
		t.Fatalf("result arrays must be non-nil: %#v", runs[0].Result)
	}
	secondApplicationID := ulid.Make().Prefixed("app")
	defer db.Where("application_id = ?", secondApplicationID).Delete(&model.SCIMSyncRun{})
	defer db.Where("application_id = ?", secondApplicationID).Delete(&model.SCIMConfiguration{})
	defer db.Where("application_id = ?", secondApplicationID).Delete(&model.ServiceProvider{})
	secondNextSyncAt := time.Now().UTC().Add(time.Hour)
	if err := db.Create(&model.ServiceProvider{ApplicationID: secondApplicationID, Profile: model.ProfileAWSIdentityCenter, EntityID: "urn:test:" + secondApplicationID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SCIMConfiguration{
		ApplicationID:  secondApplicationID,
		Endpoint:       "https://scim.us-west-2.amazonaws.com/test/scim/v2",
		AccessToken:    "secret",
		Enabled:        true,
		SyncInterval:   model.SCIMSyncInterval1Hour,
		NextSyncAt:     &secondNextSyncAt,
		LastSyncStatus: model.SCIMSyncStatusNever,
	}).Error; err != nil {
		t.Fatal(err)
	}

	const enqueueCount = 10
	queuedRuns := make(chan SCIMSyncRunView, enqueueCount)
	errorsFound := make(chan error, enqueueCount)
	var waitGroup sync.WaitGroup
	for range enqueueCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			run, _, err := EnqueueSCIMSync(context.Background(), applicationID, model.SCIMSyncTriggerManual)
			if err != nil {
				errorsFound <- err
				return
			}
			queuedRuns <- run
		}()
	}
	waitGroup.Wait()
	close(queuedRuns)
	close(errorsFound)
	for err := range errorsFound {
		t.Fatal(err)
	}
	queuedID := ""
	for run := range queuedRuns {
		if queuedID == "" {
			queuedID = run.ID
		}
		if run.ID != queuedID {
			t.Fatalf("concurrent enqueue created multiple runs: %s and %s", queuedID, run.ID)
		}
	}
	var activeRuns int64
	if err := db.Model(&model.SCIMSyncRun{}).Where("application_id = ? AND status IN ?", applicationID, []model.SCIMSyncRunStatus{
		model.SCIMSyncRunStatusQueued,
		model.SCIMSyncRunStatusRunning,
	}).Count(&activeRuns).Error; err != nil {
		t.Fatal(err)
	}
	if activeRuns != 1 {
		t.Fatalf("active runs = %d", activeRuns)
	}
	if err := DeleteSCIMConfiguration(applicationID); !errors.Is(err, ErrSCIMSyncInProgress) {
		t.Fatalf("delete with active run error = %v", err)
	}
	secondRun, created, err := EnqueueSCIMSync(context.Background(), secondApplicationID, model.SCIMSyncTriggerManual)
	if err != nil || !created {
		t.Fatalf("second application run = %#v, created = %t, error = %v", secondRun, created, err)
	}

	claimedRuns := make(chan model.SCIMSyncRun, 2)
	claimErrors := make(chan error, 2)
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			claimed, found, err := claimSCIMSyncRun(context.Background(), time.Now().UTC())
			if err != nil {
				claimErrors <- err
				return
			}
			if !found {
				claimErrors <- errors.New("no run claimed")
				return
			}
			claimedRuns <- claimed
		}()
	}
	waitGroup.Wait()
	close(claimedRuns)
	close(claimErrors)
	for err := range claimErrors {
		t.Fatal(err)
	}
	claimedByID := map[string]model.SCIMSyncRun{}
	for claimed := range claimedRuns {
		claimedByID[claimed.ID] = claimed
	}
	if len(claimedByID) != 2 {
		t.Fatalf("claimed runs = %#v", claimedByID)
	}
	recoverable, found := claimedByID[queuedID]
	if !found {
		t.Fatalf("current application run %s was not claimed", queuedID)
	}
	if err := finishSCIMSyncRun(claimedByID[secondRun.ID], SCIMSyncResult{}, nil); err != nil {
		t.Fatal(err)
	}
	expiredLease := time.Now().UTC().Add(-time.Minute)
	if err := db.Model(&model.SCIMSyncRun{}).Where("id = ?", recoverable.ID).Update("lease_expires_at", expiredLease).Error; err != nil {
		t.Fatal(err)
	}
	recovered, found, err := claimSCIMSyncRun(context.Background(), time.Now().UTC())
	if err != nil || !found || recovered.ID != recoverable.ID {
		t.Fatalf("recovered run = %#v, found = %t, error = %v", recovered, found, err)
	}
	if err := finishSCIMSyncRun(recovered, SCIMSyncResult{}, nil); err != nil {
		t.Fatal(err)
	}

	dueAt := time.Now().UTC().Add(-2 * time.Hour)
	if err := db.Model(&model.SCIMConfiguration{}).Where("application_id = ?", applicationID).Update("next_sync_at", dueAt).Error; err != nil {
		t.Fatal(err)
	}
	if err := enqueueDueSCIMSyncs(context.Background(), time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	runs, err = ListSCIMSyncRuns(applicationID, 20)
	if err != nil {
		t.Fatal(err)
	}
	if runs[0].Trigger != model.SCIMSyncTriggerScheduled || runs[0].Status != model.SCIMSyncRunStatusQueued {
		t.Fatalf("scheduled run = %#v", runs[0])
	}
	var configuration model.SCIMConfiguration
	if err := db.Where("application_id = ?", applicationID).First(&configuration).Error; err != nil {
		t.Fatal(err)
	}
	if configuration.NextSyncAt == nil || !configuration.NextSyncAt.After(time.Now().UTC()) {
		t.Fatalf("next sync = %v", configuration.NextSyncAt)
	}
}
