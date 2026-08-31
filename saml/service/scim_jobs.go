package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime/debug"
	"time"

	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

const (
	scimWorkerPollInterval  = time.Second
	scimScheduleSweep       = 30 * time.Second
	scimRunTimeout          = 30 * time.Minute
	scimRunLease            = 35 * time.Minute
	scimSyncHistoryLimit    = 50
	scimDefaultSyncInterval = model.SCIMSyncInterval1Hour
)

type SCIMSyncRunView struct {
	ID            string                  `json:"id"`
	ApplicationID string                  `json:"application_id"`
	Trigger       model.SCIMSyncTrigger   `json:"trigger"`
	Status        model.SCIMSyncRunStatus `json:"status"`
	Error         string                  `json:"error"`
	RequestedAt   time.Time               `json:"requested_at"`
	StartedAt     *time.Time              `json:"started_at"`
	CompletedAt   *time.Time              `json:"completed_at"`
	Result        SCIMSyncResult          `json:"result"`
}

func scimSyncIntervalDuration(interval model.SCIMSyncInterval) (time.Duration, error) {
	durations := map[model.SCIMSyncInterval]time.Duration{
		model.SCIMSyncInterval5Minutes:  5 * time.Minute,
		model.SCIMSyncInterval15Minutes: 15 * time.Minute,
		model.SCIMSyncInterval30Minutes: 30 * time.Minute,
		model.SCIMSyncInterval1Hour:     time.Hour,
		model.SCIMSyncInterval6Hours:    6 * time.Hour,
		model.SCIMSyncIntervalDaily:     24 * time.Hour,
	}
	duration, valid := durations[interval]
	if !valid {
		return 0, fmt.Errorf("%w: sync interval must be one of 5m, 15m, 30m, 1h, 6h, or 24h", ErrInvalidSCIMConfiguration)
	}
	return duration, nil
}

func advanceSCIMSchedule(scheduledAt time.Time, interval time.Duration, now time.Time) time.Time {
	if scheduledAt.After(now) {
		return scheduledAt
	}
	steps := now.Sub(scheduledAt)/interval + 1
	return scheduledAt.Add(steps * interval)
}

func EnqueueSCIMSync(ctx context.Context, applicationID string, trigger model.SCIMSyncTrigger) (SCIMSyncRunView, bool, error) {
	if trigger != model.SCIMSyncTriggerManual && trigger != model.SCIMSyncTriggerScheduled {
		return SCIMSyncRunView{}, false, fmt.Errorf("invalid SCIM sync trigger %q", trigger)
	}
	if err := requireAWSProfile(applicationID); err != nil {
		return SCIMSyncRunView{}, false, err
	}

	var run model.SCIMSyncRun
	created := false
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", "sentinel-scim-queue:"+applicationID).Error; err != nil {
			return fmt.Errorf("lock SCIM queue: %w", err)
		}
		var configuration model.SCIMConfiguration
		if err := tx.Where("application_id = ?", applicationID).First(&configuration).Error; err != nil {
			return err
		}
		if !configuration.Enabled {
			return fmt.Errorf("%w: provisioning is disabled", ErrInvalidSCIMConfiguration)
		}

		if trigger == model.SCIMSyncTriggerScheduled {
			if configuration.NextSyncAt == nil || configuration.NextSyncAt.After(time.Now().UTC()) {
				return nil
			}
			duration, err := scimSyncIntervalDuration(configuration.SyncInterval)
			if err != nil {
				return err
			}
			nextSyncAt := advanceSCIMSchedule(*configuration.NextSyncAt, duration, time.Now().UTC())
			if err := tx.Model(&configuration).Update("next_sync_at", nextSyncAt).Error; err != nil {
				return err
			}
		}

		lookup := tx.Where("application_id = ? AND status IN ?", applicationID, []model.SCIMSyncRunStatus{
			model.SCIMSyncRunStatusQueued,
			model.SCIMSyncRunStatusRunning,
		}).Order("requested_at DESC").Limit(1).Find(&run)
		if lookup.Error != nil {
			return lookup.Error
		}
		if lookup.RowsAffected > 0 {
			return nil
		}

		run = model.SCIMSyncRun{
			ID:               ulid.Make().Prefixed("scimsync"),
			ApplicationID:    applicationID,
			Trigger:          trigger,
			Status:           model.SCIMSyncRunStatusQueued,
			SkippedUsers:     model.SCIMSkippedUsers{},
			ValidationErrors: model.SCIMValidationErrors{},
		}
		if err := tx.Create(&run).Error; err != nil {
			return err
		}
		created = true
		return tx.Model(&model.SCIMConfiguration{}).Where("application_id = ?", applicationID).Updates(map[string]any{
			"last_sync_status": model.SCIMSyncStatusQueued,
			"last_sync_error":  "",
		}).Error
	})
	if err != nil {
		return SCIMSyncRunView{}, false, err
	}
	if run.ID == "" {
		return SCIMSyncRunView{}, false, nil
	}
	return scimSyncRunView(run), created, nil
}

func GetSCIMSyncRun(applicationID, runID string) (SCIMSyncRunView, error) {
	var run model.SCIMSyncRun
	if err := database.DB.Where("application_id = ? AND id = ?", applicationID, runID).First(&run).Error; err != nil {
		return SCIMSyncRunView{}, err
	}
	return scimSyncRunView(run), nil
}

func ListSCIMSyncRuns(applicationID string, limit int) ([]SCIMSyncRunView, error) {
	if limit <= 0 || limit > scimSyncHistoryLimit {
		limit = scimSyncHistoryLimit
	}
	var runs []model.SCIMSyncRun
	if err := database.DB.Where("application_id = ?", applicationID).Order("requested_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, err
	}
	views := make([]SCIMSyncRunView, 0, len(runs))
	for _, run := range runs {
		views = append(views, scimSyncRunView(run))
	}
	return views, nil
}

func scimSyncRunView(run model.SCIMSyncRun) SCIMSyncRunView {
	completedAt := time.Time{}
	if run.CompletedAt != nil {
		completedAt = *run.CompletedAt
	}
	skippedUsers := make([]SCIMSkippedUser, len(run.SkippedUsers))
	copy(skippedUsers, run.SkippedUsers)
	validationErrors := make([]string, len(run.ValidationErrors))
	copy(validationErrors, run.ValidationErrors)
	return SCIMSyncRunView{
		ID:            run.ID,
		ApplicationID: run.ApplicationID,
		Trigger:       run.Trigger,
		Status:        run.Status,
		Error:         run.Error,
		RequestedAt:   run.RequestedAt,
		StartedAt:     run.StartedAt,
		CompletedAt:   run.CompletedAt,
		Result: SCIMSyncResult{
			DesiredUsers:       run.DesiredUsers,
			DesiredGroups:      run.DesiredGroups,
			UsersCreated:       run.UsersCreated,
			UsersUpdated:       run.UsersUpdated,
			UsersDeactivated:   run.UsersDeactivated,
			GroupsCreated:      run.GroupsCreated,
			GroupsUpdated:      run.GroupsUpdated,
			MembershipsAdded:   run.MembershipsAdded,
			MembershipsRemoved: run.MembershipsRemoved,
			SkippedUsers:       skippedUsers,
			ValidationErrors:   validationErrors,
			CompletedAt:        completedAt,
		},
	}
}

func StartSCIMSyncWorker(ctx context.Context) {
	if err := initializeSCIMSchedules(); err != nil {
		logger.SugarLogger.Errorf("initialize SCIM schedules: %v", err)
	}
	go runSCIMSyncWorker(ctx)
}

func runSCIMSyncWorker(ctx context.Context) {
	if !waitForSCIMWorker(ctx, randomSCIMJitter(2*time.Second)) {
		return
	}
	nextScheduleSweep := time.Time{}
	for {
		now := time.Now().UTC()
		if !now.Before(nextScheduleSweep) {
			if err := enqueueDueSCIMSyncs(ctx, now); err != nil {
				logger.SugarLogger.Errorf("enqueue scheduled SCIM syncs: %v", err)
			}
			nextScheduleSweep = now.Add(scimScheduleSweep + randomSCIMJitter(10*time.Second))
		}

		run, found, err := claimSCIMSyncRun(ctx, now)
		if err != nil {
			logger.SugarLogger.Errorf("claim SCIM sync run: %v", err)
		} else if found {
			processSCIMSyncRun(ctx, run)
			continue
		}
		if !waitForSCIMWorker(ctx, scimWorkerPollInterval+randomSCIMJitter(time.Second)) {
			return
		}
	}
}

func initializeSCIMSchedules() error {
	var configurations []model.SCIMConfiguration
	if err := database.DB.Find(&configurations).Error; err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, configuration := range configurations {
		interval := configuration.SyncInterval
		duration, err := scimSyncIntervalDuration(interval)
		updates := map[string]any{}
		if err != nil {
			interval = scimDefaultSyncInterval
			duration, _ = scimSyncIntervalDuration(interval)
			updates["sync_interval"] = interval
		}
		if configuration.Enabled && configuration.NextSyncAt == nil {
			updates["next_sync_at"] = now.Add(duration)
		}
		if !configuration.Enabled && configuration.NextSyncAt != nil {
			updates["next_sync_at"] = nil
		}
		if len(updates) == 0 {
			continue
		}
		if err := database.DB.Model(&configuration).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func enqueueDueSCIMSyncs(ctx context.Context, now time.Time) error {
	var configurations []model.SCIMConfiguration
	if err := database.DB.WithContext(ctx).
		Where("enabled = ? AND next_sync_at IS NOT NULL AND next_sync_at <= ?", true, now).
		Order("next_sync_at ASC").Limit(100).Find(&configurations).Error; err != nil {
		return err
	}
	for _, configuration := range configurations {
		if _, _, err := EnqueueSCIMSync(ctx, configuration.ApplicationID, model.SCIMSyncTriggerScheduled); err != nil {
			logger.SugarLogger.Errorf("enqueue scheduled SCIM sync for application %s: %v", configuration.ApplicationID, err)
		}
	}
	return nil
}

func claimSCIMSyncRun(ctx context.Context, now time.Time) (model.SCIMSyncRun, bool, error) {
	if err := database.DB.WithContext(ctx).Model(&model.SCIMSyncRun{}).
		Where("status = ? AND lease_expires_at < ?", model.SCIMSyncRunStatusRunning, now).
		Updates(map[string]any{
			"status":           model.SCIMSyncRunStatusQueued,
			"started_at":       nil,
			"lease_expires_at": nil,
			"error":            "worker lease expired; retrying",
		}).Error; err != nil {
		return model.SCIMSyncRun{}, false, err
	}

	startedAt := now
	leaseExpiresAt := now.Add(scimRunLease)
	var run model.SCIMSyncRun
	result := database.DB.WithContext(ctx).Raw(`
		WITH candidate AS (
			SELECT id
			FROM saml_scim_sync_run
			WHERE status = ?
			ORDER BY requested_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE saml_scim_sync_run AS run
		SET status = ?, started_at = ?, lease_expires_at = ?, error = ''
		FROM candidate
		WHERE run.id = candidate.id
		RETURNING run.*
	`, model.SCIMSyncRunStatusQueued, model.SCIMSyncRunStatusRunning, startedAt, leaseExpiresAt).Scan(&run)
	if result.Error != nil {
		return model.SCIMSyncRun{}, false, result.Error
	}
	if result.RowsAffected == 0 || run.ID == "" {
		return model.SCIMSyncRun{}, false, nil
	}
	return run, true, nil
}

func processSCIMSyncRun(workerContext context.Context, run model.SCIMSyncRun) {
	var result SCIMSyncResult
	var syncErr error
	defer func() {
		if recovered := recover(); recovered != nil {
			syncErr = fmt.Errorf("SCIM sync worker panic: %v", recovered)
			logger.SugarLogger.Errorf("%v\n%s", syncErr, debug.Stack())
		}
		if err := finishSCIMSyncRun(run, result, syncErr); err != nil {
			logger.SugarLogger.Errorf("finish SCIM sync run %s: %v", run.ID, err)
		}
	}()

	ctx, cancel := context.WithTimeout(workerContext, scimRunTimeout)
	defer cancel()
	if err := database.DB.Model(&model.SCIMConfiguration{}).Where("application_id = ?", run.ApplicationID).Updates(map[string]any{
		"last_sync_status": model.SCIMSyncStatusRunning,
		"last_sync_error":  "",
	}).Error; err != nil {
		syncErr = fmt.Errorf("mark SCIM configuration running: %w", err)
		return
	}
	result, syncErr = executeSCIMSync(ctx, run.ApplicationID)
}

func finishSCIMSyncRun(run model.SCIMSyncRun, result SCIMSyncResult, syncErr error) error {
	completedAt := time.Now().UTC()
	status := model.SCIMSyncRunStatusSucceeded
	configurationStatus := model.SCIMSyncStatusSucceeded
	errorText := ""
	if syncErr != nil {
		status = model.SCIMSyncRunStatusFailed
		configurationStatus = model.SCIMSyncStatusFailed
		errorText = syncErr.Error()
		if len(errorText) > 1000 {
			errorText = errorText[:1000]
		}
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"status":              status,
			"desired_users":       result.DesiredUsers,
			"desired_groups":      result.DesiredGroups,
			"users_created":       result.UsersCreated,
			"users_updated":       result.UsersUpdated,
			"users_deactivated":   result.UsersDeactivated,
			"groups_created":      result.GroupsCreated,
			"groups_updated":      result.GroupsUpdated,
			"memberships_added":   result.MembershipsAdded,
			"memberships_removed": result.MembershipsRemoved,
			"skipped_users":       model.SCIMSkippedUsers(result.SkippedUsers),
			"validation_errors":   model.SCIMValidationErrors(result.ValidationErrors),
			"error":               errorText,
			"completed_at":        completedAt,
			"lease_expires_at":    nil,
		}
		if err := tx.Model(&model.SCIMSyncRun{}).Where("id = ?", run.ID).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&model.SCIMConfiguration{}).Where("application_id = ?", run.ApplicationID).Updates(map[string]any{
			"last_sync_at":     completedAt,
			"last_sync_status": configurationStatus,
			"last_sync_error":  errorText,
		}).Error
	})
}

func randomSCIMJitter(maximum time.Duration) time.Duration {
	if maximum <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(maximum)))
}

func waitForSCIMWorker(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
