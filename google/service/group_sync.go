package service

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
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
		Email string `json:"email"`
	} `json:"email_auth"`
	User *struct {
		Email string `json:"email"`
	} `json:"user"`
}

func getGroupMembers(groupID string) ([]coreGroupMember, error) {
	var rows []coreGroupMember
	if err := sentinel.Get("/api/groups/"+groupID+"/members", &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// resolveEntityEmail returns the entity's login email (email_auth, falling back
// to the user profile email). Empty string when the entity has no email — e.g.
// a service account — which the caller skips.
func resolveEntityEmail(entityID string) (string, error) {
	var e coreEntity
	if err := sentinel.Get("/api/core/entity/"+entityID, &e); err != nil {
		return "", err
	}
	if e.EmailAuth.Email != "" {
		return e.EmailAuth.Email, nil
	}
	if e.User != nil {
		return e.User.Email, nil
	}
	return "", nil
}

// reconcileBinding brings one Google Group's MEMBER-role membership into
// agreement with its Sentinel group. The Google Group's role=MEMBER set is the
// sync's authoritative state: anything manually added is OWNER/MANAGER and is
// never touched. Adds are skipped when the user is already present in any role.
func reconcileBinding(ctx context.Context, b model.GroupGoogleBinding) error {
	members, err := getGroupMembers(b.GroupID)
	if err != nil {
		return fmt.Errorf("fetch sentinel members for group %s: %w", b.GroupID, err)
	}

	desired := make(map[string]struct{}, len(members))
	for _, m := range members {
		email, err := resolveEntityEmail(m.EntityID)
		if err != nil {
			logger.SugarLogger.Errorf("google sync: resolve email for entity %s: %v", m.EntityID, err)
			continue
		}
		if email == "" {
			continue
		}
		desired[strings.ToLower(email)] = struct{}{}
	}

	actual, err := listGroupMembers(ctx, b.GoogleGroupEmail)
	if err != nil {
		return err
	}
	// present = members in any role (skip ADDs for these); managed = role=MEMBER
	// only (the only rows the sync may DELETE).
	present := make(map[string]struct{}, len(actual))
	managed := make(map[string]struct{}, len(actual))
	for _, a := range actual {
		le := strings.ToLower(a.Email)
		present[le] = struct{}{}
		if a.Role == "MEMBER" {
			managed[le] = struct{}{}
		}
	}

	for email := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, ok := present[email]; ok {
			continue
		}
		if err := insertMember(ctx, b.GoogleGroupEmail, email); err != nil {
			logger.SugarLogger.Errorf("google sync: %v", err)
			continue
		}
		logger.SugarLogger.Infof("google sync: added %s to %s", email, b.GoogleGroupEmail)
	}

	var toRemove []string
	for email := range managed {
		if _, ok := desired[email]; ok {
			continue
		}
		toRemove = append(toRemove, email)
	}
	if len(toRemove) > config.GoogleSyncMaxRemovals {
		logger.SugarLogger.Errorf("google sync: refusing to remove %d members from %s (exceeds GOOGLE_SYNC_MAX_REMOVALS=%d); skipping removals for this group", len(toRemove), b.GoogleGroupEmail, config.GoogleSyncMaxRemovals)
		return nil
	}
	for _, email := range toRemove {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := deleteMember(ctx, b.GoogleGroupEmail, email); err != nil {
			logger.SugarLogger.Errorf("google sync: %v", err)
			continue
		}
		logger.SugarLogger.Infof("google sync: removed %s from %s", email, b.GoogleGroupEmail)
	}
	return nil
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
		if err := reconcileBinding(ctx, b); err != nil {
			logger.SugarLogger.Errorf("google sync: reconcile failed for group=%s google=%s: %v", b.GroupID, b.GoogleGroupEmail, err)
		}
	}
	return nil
}

// sweepRunning serializes sweeps: a trigger that arrives while one is in flight
// is dropped (not queued). Safe because every sweep re-reads live state.
var sweepRunning atomic.Bool

func runSweep() {
	if directorySvc == nil {
		logger.SugarLogger.Debugln("google sync: skipping sweep, sync disabled")
		return
	}
	if !sweepRunning.CompareAndSwap(false, true) {
		logger.SugarLogger.Infoln("google sync: sweep already running, skipping this trigger")
		return
	}
	defer sweepRunning.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	logger.SugarLogger.Infoln("google sync: starting reconcile sweep")
	if err := ReconcileAll(ctx); err != nil {
		logger.SugarLogger.Errorf("google sync: sweep failed: %v", err)
		return
	}
	logger.SugarLogger.Infoln("google sync: reconcile sweep complete")
}

// TriggerReconcile kicks a sweep in the background and returns immediately.
func TriggerReconcile() {
	go runSweep()
}

// StartReconcileCron runs a periodic sweep on config.GoogleSyncInterval. A
// non-positive interval (or disabled sync) turns the cron off.
func StartReconcileCron() {
	if directorySvc == nil {
		logger.SugarLogger.Infoln("google sync: cron disabled (sync not configured)")
		return
	}
	interval := config.GoogleSyncInterval
	if interval <= 0 {
		logger.SugarLogger.Infof("google sync: cron disabled (interval=%v)", interval)
		return
	}
	logger.SugarLogger.Infof("google sync: cron enabled, interval=%v", interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			runSweep()
		}
	}()
}
