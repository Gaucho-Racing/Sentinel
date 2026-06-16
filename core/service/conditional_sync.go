package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
)

// Per-entity reconciles serialize via a syncJobMap keyed by entity ID; the
// full sweep is its own singleton syncJob. Cancel-and-restart semantics:
// a newer trigger for the same entity (or a newer full-sweep trigger)
// cancels the in-flight run and replaces it. Reconcile is idempotent — a
// cancelled run leaves the DB in a consistent (if partial) state and the
// next run catches up.
var (
	conditionalEntityJobs syncJobMap
	conditionalSweepJob   syncJob
)

// maxFixedPointRounds caps the per-entity convergence loop. With cycle
// detection at binding-creation time, the dependency graph is a DAG and
// any well-formed input converges in O(depth) rounds. 16 is wildly more
// than any realistic binding depth — exceeding it is a sign of either a
// bug in the cycle check or a race between concurrent binding edits.
const maxFixedPointRounds = 16

// ReconcileConditionalForEntity reconciles an entity's CONDITIONAL-sourced
// group memberships against the current conditional bindings. Returns
// immediately after scheduling the work; completion is asynchronous and
// failures are logged, not propagated. A subsequent call for the same
// entity cancels the in-flight run and starts a new one.
//
// The fixed-point loop inside handles transitive composition: when adding
// an entity to group X newly satisfies another group's binding (one that
// requires X), the NEXT iteration sees the new membership and adds them
// to that group too. Continues until a round produces no changes.
//
// The error return is kept for caller ergonomics but is always nil.
func ReconcileConditionalForEntity(entityID string) error {
	conditionalEntityJobs.Start(entityID, func(ctx context.Context) {
		if err := reconcileConditionalForEntityCtx(ctx, entityID); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.SugarLogger.Debugf("conditional sync: per-entity run for %s cancelled by newer event", entityID)
				return
			}
			logger.SugarLogger.Errorf("conditional sync: reconcile failed for %s: %v", entityID, err)
		}
	})
	return nil
}

func reconcileConditionalForEntityCtx(ctx context.Context, entityID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Snapshot bindings + allowed_sources once per reconcile call — they
	// don't change mid-loop, so we don't need to re-fetch every round.
	// Memberships ARE re-read each round since the loop itself mutates them.
	bindings, err := GetAllConditionalBindings()
	if err != nil {
		return fmt.Errorf("load bindings: %w", err)
	}
	bindingsByGroup := make(map[string][]model.GroupConditionalBinding, len(bindings))
	for _, b := range bindings {
		bindingsByGroup[b.GroupID] = append(bindingsByGroup[b.GroupID], b)
	}

	// conditionalEnabled gates which groups can receive CONDITIONAL
	// memberships from this sync — mirrors discordEnabled in the Discord
	// sync. Bindings on a group that has revoked CONDITIONAL from its
	// allowed_sources are skipped, so cascade-removed members can't get
	// re-added on the next reconcile.
	var allGroups []model.Group
	if err := database.DB.Find(&allGroups).Error; err != nil {
		return fmt.Errorf("load groups: %w", err)
	}
	conditionalEnabled := make(map[string]bool, len(allGroups))
	for _, g := range allGroups {
		for _, src := range g.AllowedSources {
			if src == string(model.GroupMemberSourceConditional) {
				conditionalEnabled[g.ID] = true
				break
			}
		}
	}

	for round := 0; round < maxFixedPointRounds; round++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		memberships, err := GetMembershipsForEntity(entityID, "")
		if err != nil {
			return fmt.Errorf("fetch memberships: %w", err)
		}

		// Set of every group the entity is in (any source) — used to evaluate
		// bindings. CONDITIONAL memberships count too, which is what enables
		// transitive composition.
		memberGroups := make(map[string]struct{}, len(memberships))
		// Map of group → source for the entity's existing memberships, used
		// to scope our deletes to CONDITIONAL-only (so we never accidentally
		// remove a DIRECT or DISCORD-sourced row).
		sourceByGroup := make(map[string]string, len(memberships))
		for _, m := range memberships {
			memberGroups[m.GroupID] = struct{}{}
			sourceByGroup[m.GroupID] = m.Source
		}
		memberGroupIDs := make([]string, 0, len(memberGroups))
		for g := range memberGroups {
			memberGroupIDs = append(memberGroupIDs, g)
		}

		changes := 0
		for parentGroup, bs := range bindingsByGroup {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Skip groups that don't allow CONDITIONAL — bindings on them are
			// orphans (admin removed CONDITIONAL from allowed_sources without
			// deleting the bindings). cascadeRemovedSources already stripped
			// their existing CONDITIONAL members on the allowed_sources edit.
			if !conditionalEnabled[parentGroup] {
				continue
			}
			satisfied := EvaluateConditionalMembership(bs, memberGroupIDs)
			_, isMember := memberGroups[parentGroup]
			currentSource := sourceByGroup[parentGroup]

			if satisfied && !isMember {
				if _, err := CreateGroupMember(model.GroupMember{
					GroupID:  parentGroup,
					EntityID: entityID,
					Source:   string(model.GroupMemberSourceConditional),
					AddedBy:  SentinelServiceAccountName,
				}); err != nil {
					logger.SugarLogger.Errorf("conditional sync: failed to add %s to %s: %v", entityID, parentGroup, err)
					continue
				}
				logger.SugarLogger.Infof("conditional sync: added entity %s to group %s (CONDITIONAL)", entityID, parentGroup)
				changes++
			} else if !satisfied && isMember && currentSource == string(model.GroupMemberSourceConditional) {
				if err := DeleteGroupMember(parentGroup, entityID, string(model.GroupMemberSourceConditional)); err != nil {
					logger.SugarLogger.Errorf("conditional sync: failed to remove %s from %s: %v", entityID, parentGroup, err)
					continue
				}
				logger.SugarLogger.Infof("conditional sync: removed entity %s from group %s (CONDITIONAL)", entityID, parentGroup)
				changes++
			}
		}

		if changes == 0 {
			return nil
		}
	}

	logger.SugarLogger.Errorf("conditional sync: entity %s did not converge after %d rounds", entityID, maxFixedPointRounds)
	return fmt.Errorf("did not converge after %d rounds", maxFixedPointRounds)
}

// TriggerReconcileAllConditional schedules a full reconcile across every
// entity in the system. Used on binding mutations and on the periodic cron.
// Subsequent calls (e.g. a second binding mutation arriving while a sweep is
// running) cancel the in-flight sweep and start a new one with the latest
// binding set.
func TriggerReconcileAllConditional() {
	conditionalSweepJob.Start(func(ctx context.Context) {
		if err := reconcileAllConditionalCtx(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.SugarLogger.Debugf("conditional sync: full sweep cancelled by newer trigger")
				return
			}
			logger.SugarLogger.Errorf("conditional sync: full sweep failed: %v", err)
		}
	})
}

// reconcileAllConditionalCtx walks every entity in the DB and reconciles
// their conditional memberships. Could be optimized to only walk entities
// who have a chance of matching some binding, but at typical org scale a
// full scan is fine and simpler. ctx is checked between entities.
func reconcileAllConditionalCtx(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var entityIDs []string
	if err := database.DB.
		Model(&model.Entity{}).
		Pluck("id", &entityIDs).Error; err != nil {
		return fmt.Errorf("list entities: %w", err)
	}
	if len(entityIDs) == 0 {
		return nil
	}

	logger.SugarLogger.Infof("conditional sync: starting full sweep over %d entities", len(entityIDs))
	for i, entityID := range entityIDs {
		if err := ctx.Err(); err != nil {
			logger.SugarLogger.Infof("conditional sync: sweep cancelled after %d/%d entities", i, len(entityIDs))
			return err
		}
		if err := reconcileConditionalForEntityCtx(ctx, entityID); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			logger.SugarLogger.Errorf("conditional sync: reconcile failed for entity=%s: %v", entityID, err)
		}
	}
	logger.SugarLogger.Infof("conditional sync: full sweep complete")
	return nil
}

// StartReconcileConditionalCron spawns a background goroutine that ticks
// TriggerReconcileAllConditional on config.ConditionalSyncInterval. Same
// pattern as the discord sync cron: safety net for any drift the event
// stream misses. Non-positive interval disables the cron.
func StartReconcileConditionalCron() {
	interval := config.ConditionalSyncInterval
	if interval <= 0 {
		logger.SugarLogger.Infof("conditional sync: cron disabled (interval=%v)", interval)
		return
	}
	logger.SugarLogger.Infof("conditional sync: cron enabled, interval=%v", interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			logger.SugarLogger.Debugf("conditional sync: cron tick, kicking full sweep")
			TriggerReconcileAllConditional()
		}
	}()
}
