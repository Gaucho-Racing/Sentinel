package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/discord/config"
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
	allMemberships, err := getEntityMemberships(entity.ID)
	if err != nil {
		return fmt.Errorf("fetch current memberships: %w", err)
	}

	desiredSet := toSet(desired)
	// allMemberSet = every group the entity is in via any source. Used
	// for the "should I add?" check so we skip groups where the user is
	// already a member via DIRECT or CONDITIONAL — otherwise the
	// (group_id, entity_id) primary key trips on the add.
	allMemberSet := make(map[string]struct{}, len(allMemberships))
	// discordMemberSet = DISCORD-sourced rows only. The delete loop
	// considers these alone, scoped to source=DISCORD on the delete
	// call so we never touch DIRECT/CONDITIONAL rows.
	discordMemberSet := make(map[string]struct{}, len(allMemberships))
	for _, m := range allMemberships {
		allMemberSet[m.GroupID] = struct{}{}
		if m.Source == "DISCORD" {
			discordMemberSet[m.GroupID] = struct{}{}
		}
	}

	for groupID := range desiredSet {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, already := allMemberSet[groupID]; already {
			continue
		}
		if err := addDiscordGroupMember(groupID, entity.ID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to add %s to %s: %v", entity.ID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: added entity %s to group %s (DISCORD)", entity.ID, groupID)
	}
	for groupID := range discordMemberSet {
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

// getEntityMemberships returns every group_member row for the entity,
// regardless of source. Callers (the reconcile diff) derive two sets
// from the result: one for "is already a member via anything" (used to
// skip ADDs) and one for "is a DISCORD-sourced member" (used to drive
// DELETEs). Fetching all in one round-trip keeps the per-user reconcile
// to a single core lookup for the membership state.
func getEntityMemberships(entityID string) ([]groupMemberRow, error) {
	var rows []groupMemberRow
	if err := sentinel.Get("/api/core/entity/"+entityID+"/memberships", &rows); err != nil {
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

// StartReconcileCron spawns a background goroutine that periodically calls
// TriggerReconcileAll on config.GroupSyncInterval. Acts as a safety net for
// drift the event stream might miss: dropped gateway events, bot restarts,
// out-of-band core-side changes (e.g. group allowed_sources flips) that
// don't surface as Discord events. A non-positive interval disables the
// cron — useful in tests and one-off runs.
//
// Cancel-and-restart semantics in TriggerReconcileAll make this safe: if a
// sweep is still running when the next tick fires, the in-flight sweep is
// cancelled and replaced. The goroutine is leaked on process exit; we don't
// have graceful-shutdown plumbing for it and the OS reaps everything anyway.
func StartReconcileCron() {
	interval := config.GroupSyncInterval
	if interval <= 0 {
		logger.SugarLogger.Infof("group sync: cron disabled (interval=%v)", interval)
		return
	}
	logger.SugarLogger.Infof("group sync: cron enabled, interval=%v", interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			logger.SugarLogger.Debugf("group sync: cron tick, kicking full sweep")
			TriggerReconcileAll()
		}
	}()
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

	if err := ctx.Err(); err != nil {
		return err
	}
	memberRoles, err := fetchAllGuildMemberRoles(ctx)
	if err != nil {
		return fmt.Errorf("fetch guild members: %w", err)
	}

	logger.SugarLogger.Infof("group sync: starting full sweep over %d onboarded discord users", len(auths))
	for _, a := range auths {
		if err := ctx.Err(); err != nil {
			logger.SugarLogger.Infof("group sync: sweep cancelled mid-iteration after %d users", indexOf(auths, a))
			return err
		}
		// Absent from the authoritative member list means the user has left
		// the guild. Skip rather than strip — OnGuildMemberRemove already
		// handles leave-cleanup, and skipping avoids touching memberships if
		// the bulk fetch returned a partial picture.
		roles, present := memberRoles[a.ExternalID]
		if !present {
			logger.SugarLogger.Debugf("group sync: skipping entity=%s discord=%s, not in guild member list", a.EntityID, a.ExternalID)
			continue
		}
		if err := reconcileOneWithSnapshot(ctx, a.EntityID, roles, bindingsByGroup, discordEnabled); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			logger.SugarLogger.Errorf("group sync: reconcile failed for entity=%s discord=%s: %v", a.EntityID, a.ExternalID, err)
		}
	}
	logger.SugarLogger.Infof("group sync: full sweep complete")
	return nil
}

// fetchAllGuildMemberRoles returns an authoritative discordUserID -> role IDs
// map for the whole guild, paginated over the GuildMembers REST endpoint
// (1000 per page). The full sweep uses this instead of per-user State cache
// lookups: the State cache can hold partial members with stale or empty
// Roles, which silently yields the wrong desired set. ctx is checked between
// pages so a cancellation lands at the next page boundary.
func fetchAllGuildMemberRoles(ctx context.Context) (map[string][]string, error) {
	roles := make(map[string][]string)
	after := ""
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		members, err := Discord.GuildMembers(config.DiscordGuild, after, 1000)
		if err != nil {
			return nil, err
		}
		if len(members) == 0 {
			break
		}
		for _, m := range members {
			if m.User == nil {
				continue
			}
			roles[m.User.ID] = m.Roles
			after = m.User.ID
		}
		if len(members) < 1000 {
			break
		}
	}
	return roles, nil
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
// pre-fetched binding + allowed_sources snapshots and the user's
// authoritative Discord roles (from the sweep's bulk member fetch).
func reconcileOneWithSnapshot(ctx context.Context, entityID string, roles []string, bindingsByGroup map[string][]model.GroupDiscordRoleBinding, discordEnabled map[string]bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	desired := make(map[string]struct{})
	for groupID, bs := range bindingsByGroup {
		if !discordEnabled[groupID] {
			continue
		}
		if EvaluateDiscordMembership(bs, roles) {
			desired[groupID] = struct{}{}
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	allMemberships, err := getEntityMemberships(entityID)
	if err != nil {
		return fmt.Errorf("fetch current memberships: %w", err)
	}
	// Same two-set pattern as the per-user reconcile: skip ADDs when
	// the entity is already a member via any source; only consider
	// DISCORD rows for DELETEs.
	allMemberSet := make(map[string]struct{}, len(allMemberships))
	discordMemberSet := make(map[string]struct{}, len(allMemberships))
	for _, m := range allMemberships {
		allMemberSet[m.GroupID] = struct{}{}
		if m.Source == "DISCORD" {
			discordMemberSet[m.GroupID] = struct{}{}
		}
	}

	for groupID := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, already := allMemberSet[groupID]; already {
			continue
		}
		if err := addDiscordGroupMember(groupID, entityID); err != nil {
			logger.SugarLogger.Errorf("group sync: failed to add %s to %s: %v", entityID, groupID, err)
			continue
		}
		logger.SugarLogger.Infof("group sync: added entity %s to group %s (DISCORD)", entityID, groupID)
	}
	for groupID := range discordMemberSet {
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
