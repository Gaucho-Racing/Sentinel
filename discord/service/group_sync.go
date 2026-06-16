package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
)

// groupSummary mirrors the relevant fields of core's /groups response.
// Only the fields we need to decide whether DISCORD is an allowed source.
type groupSummary struct {
	ID             string   `json:"id"`
	AllowedSources []string `json:"allowed_sources"`
}

// groupMemberRow mirrors core/model/group.go::GroupMember on the wire.
type groupMemberRow struct {
	GroupID  string `json:"group_id"`
	EntityID string `json:"entity_id"`
	Source   string `json:"source"`
}

// Per-user reconciles are serialized via a syncJobMap keyed by Discord user
// ID. The full sweep is its own singleton syncJob. Both use cancel-and-
// restart: a new event for the same user (or a new binding mutation) cancels
// the in-flight run and replaces it with a fresh one. This works because
// every reconcile re-reads the live state — a cancelled run leaves the DB in
// a consistent (if partial) state and the next run catches up from there.
var (
	userSyncJobs syncJobMap
	sweepJob     syncJob
)

// ReconcileGroupsForDiscordUser brings a user's DISCORD-sourced group
// memberships into agreement with their current Discord role set. Returns
// immediately after scheduling the work — completion is asynchronous. A
// subsequent call for the same Discord user cancels the in-flight run and
// schedules a new one with the latest role set.
//
// The error return is kept for caller ergonomics but is always nil; failures
// inside the spawned goroutine are logged, not propagated.
func ReconcileGroupsForDiscordUser(discordUserID string, currentRoles []string) error {
	roles := append([]string(nil), currentRoles...) // defensive copy for the closure
	userSyncJobs.Start(discordUserID, func(ctx context.Context) {
		if err := reconcileGroupsForDiscordUserCtx(ctx, discordUserID, roles); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.SugarLogger.Debugf("group sync: per-user run for %s cancelled by newer event", discordUserID)
				return
			}
			logger.SugarLogger.Errorf("group sync: reconcile failed for %s: %v", discordUserID, err)
		}
	})
	return nil
}

// reconcileGroupsForDiscordUserCtx is the ctx-aware body. It checks
// ctx.Err() between every cross-service step so a cancellation lands at the
// next safe boundary; mid-HTTP-call cancellation is not attempted.
func reconcileGroupsForDiscordUserCtx(ctx context.Context, discordUserID string, currentRoles []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var entity entityResponse
	if err := sentinel.Get("/api/core/entity/external/DISCORD/"+discordUserID, &entity); err != nil {
		logger.SugarLogger.Debugf("group sync: no entity for Discord user %s: %v", discordUserID, err)
		return nil
	}
	if entity.ID == "" {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	desired, err := computeDesiredDiscordGroups(currentRoles)
	if err != nil {
		return fmt.Errorf("compute desired groups: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	current, err := getEntityDiscordMemberships(entity.ID)
	if err != nil {
		return fmt.Errorf("fetch current memberships: %w", err)
	}

	desiredSet := toSet(desired)
	currentSet := make(map[string]struct{}, len(current))
	for _, m := range current {
		currentSet[m.GroupID] = struct{}{}
	}

	for groupID := range desiredSet {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, already := currentSet[groupID]; already {
			continue
		}
		if err := addDiscordGroupMember(groupID, entity.ID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to add %s to %s: %v", entity.ID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: added entity %s to group %s (DISCORD)", entity.ID, groupID)
	}
	for groupID := range currentSet {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, keep := desiredSet[groupID]; keep {
			continue
		}
		if err := removeDiscordGroupMember(groupID, entity.ID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to remove %s from %s: %v", entity.ID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: removed entity %s from group %s (DISCORD)", entity.ID, groupID)
	}
	return nil
}

// computeDesiredDiscordGroups returns the set of group IDs the user should
// belong to via DISCORD: bindings that match their roles, intersected with
// groups whose allowed_sources still includes DISCORD. The intersection
// step is what keeps orphaned bindings (group revoked DISCORD source but
// bindings weren't cleaned) from re-adding cascade-removed members.
func computeDesiredDiscordGroups(userRoles []string) ([]string, error) {
	eligible, err := GetEligibleGroupsForUserRoles(userRoles)
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return nil, nil
	}
	var groups []groupSummary
	if err := sentinel.Get("/api/groups", &groups); err != nil {
		return nil, err
	}
	discordEnabled := make(map[string]bool, len(groups))
	for _, g := range groups {
		for _, src := range g.AllowedSources {
			if src == "DISCORD" {
				discordEnabled[g.ID] = true
				break
			}
		}
	}
	desired := make([]string, 0, len(eligible))
	for _, gid := range eligible {
		if discordEnabled[gid] {
			desired = append(desired, gid)
		}
	}
	return desired, nil
}

func getEntityDiscordMemberships(entityID string) ([]groupMemberRow, error) {
	var rows []groupMemberRow
	if err := sentinel.Get("/api/core/entity/"+entityID+"/memberships?source=DISCORD", &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func addDiscordGroupMember(groupID, entityID string) error {
	body := map[string]any{
		"entity_id": entityID,
		"source":    "DISCORD",
		"added_by":  "discord-sync",
	}
	return sentinel.Post("/api/groups/"+groupID+"/members", body, nil)
}

func removeDiscordGroupMember(groupID, entityID string) error {
	return sentinel.Delete("/api/groups/"+groupID+"/members/"+entityID+"?source=DISCORD", nil)
}

func toSet(ss []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		out[s] = struct{}{}
	}
	return out
}

// externalAuthRow mirrors the EntityExternalAuth fields needed to enumerate
// onboarded users for a provider.
type externalAuthRow struct {
	EntityID   string `json:"entity_id"`
	ExternalID string `json:"external_id"`
}

// TriggerReconcileAll schedules a full reconciliation sweep over every
// onboarded Discord user. A subsequent call (e.g. a second role-binding
// mutation arriving while the first sweep is still running) cancels the
// in-flight sweep and starts a new one with the latest binding state. Safe
// because the sweep re-reads bindings and allowed_sources at the top of
// every run — cancelling mid-iteration just means the next run picks up
// users that weren't yet processed.
func TriggerReconcileAll() {
	sweepJob.Start(func(ctx context.Context) {
		if err := reconcileAllOnboardedDiscordUsers(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.SugarLogger.Debugf("group sync: full sweep cancelled by newer trigger")
				return
			}
			logger.SugarLogger.Errorf("group sync: full sweep failed: %v", err)
		}
	})
}

// reconcileAllOnboardedDiscordUsers walks every Sentinel entity with a
// DISCORD external auth and reconciles their DISCORD-sourced group
// memberships against current Discord roles. Bindings and group
// allowed_sources are snapshotted once up-front so the per-user inner loop
// only touches core for memberships + diff writes. ctx is checked between
// iterations so a cancellation cuts off the sweep at the next user boundary.
func reconcileAllOnboardedDiscordUsers(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var auths []externalAuthRow
	if err := sentinel.Get("/api/core/entity/external/DISCORD", &auths); err != nil {
		return fmt.Errorf("list discord external auths: %w", err)
	}
	if len(auths) == 0 {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	allBindings, err := GetAllRoleBindings()
	if err != nil {
		return fmt.Errorf("load bindings: %w", err)
	}
	bindingsByGroup := make(map[string][]model.GroupDiscordRoleBinding, len(allBindings))
	for _, b := range allBindings {
		bindingsByGroup[b.GroupID] = append(bindingsByGroup[b.GroupID], b)
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	var groups []groupSummary
	if err := sentinel.Get("/api/groups", &groups); err != nil {
		return fmt.Errorf("load groups: %w", err)
	}
	discordEnabled := make(map[string]bool, len(groups))
	for _, g := range groups {
		for _, src := range g.AllowedSources {
			if src == "DISCORD" {
				discordEnabled[g.ID] = true
				break
			}
		}
	}

	logger.SugarLogger.Infof("group sync: starting full sweep over %d onboarded discord users", len(auths))
	for _, a := range auths {
		if err := ctx.Err(); err != nil {
			logger.SugarLogger.Infof("group sync: sweep cancelled mid-iteration after %d users", indexOf(auths, a))
			return err
		}
		if err := reconcileOneWithSnapshot(ctx, a.EntityID, a.ExternalID, bindingsByGroup, discordEnabled); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			logger.SugarLogger.Errorf("group sync: reconcile failed for entity=%s discord=%s: %v", a.EntityID, a.ExternalID, err)
		}
	}
	logger.SugarLogger.Infof("group sync: full sweep complete")
	return nil
}

// indexOf returns the index of row in auths, or -1 if not found. Used only
// for the cancellation log line so the operator can see how far the sweep
// got before being cancelled.
func indexOf(auths []externalAuthRow, row externalAuthRow) int {
	for i, a := range auths {
		if a.EntityID == row.EntityID && a.ExternalID == row.ExternalID {
			return i
		}
	}
	return -1
}

// reconcileOneWithSnapshot reconciles a single onboarded user using
// pre-fetched binding + allowed_sources snapshots. Reads current Discord
// roles via the guild-member lookup (state cache then REST fallback).
// Users no longer present in the guild are skipped rather than stripped —
// a transient lookup failure shouldn't aggressively remove memberships.
func reconcileOneWithSnapshot(ctx context.Context, entityID, discordID string, bindingsByGroup map[string][]model.GroupDiscordRoleBinding, discordEnabled map[string]bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	member, err := GetGuildMember(discordID)
	if err != nil {
		logger.SugarLogger.Debugf("group sync: skipping entity=%s discord=%s, guild member lookup failed: %v", entityID, discordID, err)
		return nil
	}

	desired := make(map[string]struct{})
	for groupID, bs := range bindingsByGroup {
		if !discordEnabled[groupID] {
			continue
		}
		if EvaluateDiscordMembership(bs, member.Roles) {
			desired[groupID] = struct{}{}
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	current, err := getEntityDiscordMemberships(entityID)
	if err != nil {
		return fmt.Errorf("fetch current memberships: %w", err)
	}
	currentSet := make(map[string]struct{}, len(current))
	for _, m := range current {
		currentSet[m.GroupID] = struct{}{}
	}

	for groupID := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, already := currentSet[groupID]; already {
			continue
		}
		if err := addDiscordGroupMember(groupID, entityID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to add %s to %s: %v", entityID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: added entity %s to group %s (DISCORD)", entityID, groupID)
	}
	for groupID := range currentSet {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, keep := desired[groupID]; keep {
			continue
		}
		if err := removeDiscordGroupMember(groupID, entityID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to remove %s from %s: %v", entityID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: removed entity %s from group %s (DISCORD)", entityID, groupID)
	}
	return nil
}
