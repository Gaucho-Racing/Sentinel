package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/google/config"
	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
	"github.com/gaucho-racing/sentinel/google/pkg/sentinel"
)

// coreGroupMember mirrors the fields of core's GroupMember we need.
type coreGroupMember struct {
	EntityID string `json:"entity_id"`
	Source   string `json:"source"`
}

// coreEntity mirrors the entity fields needed to resolve a member's email.
type coreEntity struct {
	EmailAuth struct {
		Email      string  `json:"email"`
		VerifiedAt *string `json:"verified_at"`
	} `json:"email_auth"`
}

// protectedEmailDomain addresses are team accounts the sync never removes.
const protectedEmailDomain = "@gauchoracing.com"

func isProtectedEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), protectedEmailDomain)
}

// getRetiredEmails lists addresses Sentinel users had before changing their
// email, which no account uses anymore.
func getRetiredEmails() (map[string]struct{}, error) {
	var emails []string
	if err := sentinel.Get("/api/core/internal/retired-emails", &emails); err != nil {
		return nil, err
	}
	retired := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		retired[strings.ToLower(email)] = struct{}{}
	}
	return retired, nil
}

func getGroupMembers(groupID string) ([]coreGroupMember, error) {
	var rows []coreGroupMember
	if err := sentinel.Get("/api/groups/"+groupID+"/members", &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// resolveEntityEmail returns the entity's login email and whether it is
// verified. Empty string when the entity has no email — e.g. a service
// account — which the caller skips.
func resolveEntityEmail(entityID string) (string, bool, error) {
	var e coreEntity
	if err := sentinel.Get("/api/core/entity/"+entityID, &e); err != nil {
		return "", false, err
	}
	return e.EmailAuth.Email, e.EmailAuth.VerifiedAt != nil, nil
}

// reconcileBinding brings one Google Group's MEMBER-role membership into
// agreement with its Sentinel group. The Google Group's role=MEMBER set is the
// sync's authoritative state: anything manually added is OWNER/MANAGER and is
// never touched. Adds are skipped when the user is already present in any role.
func reconcileBinding(ctx context.Context, b model.GroupGoogleBinding, allowBulkRemovals bool) error {
	if err := ensureGroupOwner(ctx, b.GoogleGroupEmail, ManagedGoogleGroupOwnerEmail); err != nil {
		return err
	}
	members, err := getGroupMembers(b.GroupID)
	if err != nil {
		return fmt.Errorf("fetch sentinel members for group %s: %w", b.GroupID, err)
	}

	// Only verified addresses are added. An unverified member's current
	// address is left in place if it is already there, so members who have not
	// verified yet keep their access instead of being removed.
	desired := make(map[string]struct{}, len(members))
	unverified := make(map[string]struct{})
	for _, m := range members {
		email, verified, err := resolveEntityEmail(m.EntityID)
		if err != nil {
			logger.SugarLogger.Errorf("google sync: resolve email for entity %s: %v", m.EntityID, err)
			continue
		}
		if email == "" {
			continue
		}
		if verified {
			desired[strings.ToLower(email)] = struct{}{}
		} else {
			unverified[strings.ToLower(email)] = struct{}{}
		}
	}

	retired, err := getRetiredEmails()
	if err != nil {
		logger.SugarLogger.Errorf("google sync: load retired emails: %v", err)
		retired = map[string]struct{}{}
	}

	actual, err := listGroupMembers(ctx, b.GoogleGroupEmail)
	if err != nil {
		return err
	}
	// present = members in any role (skip ADDs for these); managed = role=MEMBER
	// only (the only rows the sync may DELETE).
	// A retired address is removed in any role, since it can only be there
	// because its owner changed their email.
	present := make(map[string]struct{}, len(actual))
	managed := make(map[string]struct{}, len(actual))
	for _, a := range actual {
		le := strings.ToLower(a.Email)
		present[le] = struct{}{}
		if _, isRetired := retired[le]; a.Role == "MEMBER" || isRetired {
			managed[le] = struct{}{}
		}
	}

	var operationErrors []error
	for email := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, ok := present[email]; ok {
			continue
		}
		if err := insertMember(ctx, b.GoogleGroupEmail, email); err != nil {
			logger.SugarLogger.Errorf("google sync: %v", err)
			operationErrors = append(operationErrors, err)
			continue
		}
		logger.SugarLogger.Infof("google sync: added %s to %s", email, b.GoogleGroupEmail)
	}

	var toRemove []string
	for email := range managed {
		if _, ok := desired[email]; ok {
			continue
		}
		if _, ok := unverified[email]; ok {
			continue
		}
		if isProtectedEmail(email) {
			continue
		}
		toRemove = append(toRemove, email)
	}
	if !allowBulkRemovals && len(toRemove) > config.GoogleSyncMaxRemovals {
		logger.SugarLogger.Errorf("google sync: refusing to remove %d members from %s (exceeds GOOGLE_SYNC_MAX_REMOVALS=%d); skipping removals for this group", len(toRemove), b.GoogleGroupEmail, config.GoogleSyncMaxRemovals)
		return nil
	}
	for _, email := range toRemove {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := deleteMember(ctx, b.GoogleGroupEmail, email); err != nil {
			logger.SugarLogger.Errorf("google sync: %v", err)
			operationErrors = append(operationErrors, err)
			continue
		}
		logger.SugarLogger.Infof("google sync: removed %s from %s", email, b.GoogleGroupEmail)
	}
	return errors.Join(operationErrors...)
}

// ReconcileAll reconciles every binding. A failure on one binding is logged and
// does not abort the others.
func ReconcileAll(ctx context.Context) error {
	bindings, err := GetAllGoogleBindings()
	if err != nil {
		return fmt.Errorf("load bindings: %w", err)
	}
	for _, b := range bindings {
		if err := ctx.Err(); err != nil {
			return err
		}
		var reconcileErr error
		if b.Status == BindingStatusPending {
			logger.SugarLogger.Infof("google binding: starting operation=%s group=%s", b.OperationID, b.GroupID)
			reconcileErr = applyPendingGoogleBinding(ctx, b)
		} else {
			reconcileErr = reconcileBinding(ctx, b, false)
		}
		if reconcileErr != nil {
			if b.Status == BindingStatusPending {
				recordGoogleBindingSyncError(b, reconcileErr)
			}
			if errors.Is(reconcileErr, context.Canceled) {
				return reconcileErr
			}
			logger.SugarLogger.Errorf("google sync: reconcile failed for group=%s google=%s: %v", b.GroupID, b.GoogleGroupEmail, reconcileErr)
		} else if b.Status == BindingStatusPending {
			logger.SugarLogger.Infof("google binding: completed operation=%s group=%s", b.OperationID, b.GroupID)
		}
	}
	return nil
}

// sweepJob serializes sweeps with cancel-and-restart: a trigger that arrives
// while one is in flight cancels it and runs a fresh sweep with the latest
// state. Safe because every sweep re-reads live state and reconcile is
// idempotent.
var sweepJob syncJob

func runSweep() {
	if directorySvc == nil {
		logger.SugarLogger.Debugln("google sync: skipping sweep, sync disabled")
		return
	}
	sweepJob.Start(func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		logger.SugarLogger.Infoln("google sync: starting reconcile sweep")
		if err := ReconcileAll(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.SugarLogger.Debugln("google sync: sweep cancelled by newer trigger")
				return
			}
			logger.SugarLogger.Errorf("google sync: sweep failed: %v", err)
			return
		}
		logger.SugarLogger.Infoln("google sync: reconcile sweep complete")
	})
}

// TriggerReconcile kicks a sweep and returns immediately. A sweep already in
// flight is cancelled in favor of this one.
func TriggerReconcile() {
	runSweep()
}

// StartReconcileCron runs a periodic sweep on config.GoogleSyncInterval. The
// cron is off only when sync isn't configured.
func StartReconcileCron() {
	if directorySvc == nil {
		logger.SugarLogger.Infoln("google sync: cron disabled (sync not configured)")
		return
	}
	logger.SugarLogger.Infof("google sync: cron enabled, interval=%v", config.GoogleSyncInterval)
	go func() {
		ticker := time.NewTicker(config.GoogleSyncInterval)
		defer ticker.Stop()
		for range ticker.C {
			runSweep()
		}
	}()
	runSweep()
}
