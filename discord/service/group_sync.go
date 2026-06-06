package service

import (
	"fmt"
	"sync"

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

// ReconcileGroupsForDiscordUser brings a user's DISCORD-sourced group
// memberships into agreement with their current Discord role set. It is a
// no-op if the Discord user has no Sentinel entity (not onboarded).
//
// Steps:
//  1. Resolve Discord ID -> entity ID via core
//  2. Compute desired groups: bindings that match the user's roles, scoped
//     to groups that still allow DISCORD in their allowed_sources
//  3. Read the entity's current DISCORD-sourced membership rows from core
//  4. Apply the diff via core's group-member endpoints
func ReconcileGroupsForDiscordUser(discordUserID string, currentRoles []string) error {
	var entity entityResponse
	if err := sentinel.Get("/api/core/entity/external/DISCORD/"+discordUserID, &entity); err != nil {
		logger.SugarLogger.Debugf("group sync: no entity for Discord user %s: %v", discordUserID, err)
		return nil
	}
	if entity.ID == "" {
		return nil
	}

	desired, err := computeDesiredDiscordGroups(currentRoles)
	if err != nil {
		return fmt.Errorf("compute desired groups: %w", err)
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

var (
	sweepMu      sync.Mutex
	sweepRunning bool
	sweepDirty   bool
)

// TriggerReconcileAll requests a full reconciliation sweep over every
// onboarded Discord user. If a sweep is already running, the request is
// coalesced — when the current sweep finishes, it will run once more to
// pick up the latest binding state. This guarantees the final sweep
// reflects any binding mutations made during a prior sweep, without
// spawning parallel sweeps.
func TriggerReconcileAll() {
	sweepMu.Lock()
	if sweepRunning {
		sweepDirty = true
		sweepMu.Unlock()
		return
	}
	sweepRunning = true
	sweepMu.Unlock()

	go func() {
		for {
			if err := reconcileAllOnboardedDiscordUsers(); err != nil {
				logger.SugarLogger.Errorf("group sync: full sweep failed: %v", err)
			}
			sweepMu.Lock()
			if !sweepDirty {
				sweepRunning = false
				sweepMu.Unlock()
				return
			}
			sweepDirty = false
			sweepMu.Unlock()
		}
	}()
}

// reconcileAllOnboardedDiscordUsers walks every Sentinel entity with a
// DISCORD external auth and reconciles their DISCORD-sourced group
// memberships against current Discord roles. Bindings and group
// allowed_sources are snapshotted once up-front so the per-user inner loop
// only touches core for memberships + diff writes.
func reconcileAllOnboardedDiscordUsers() error {
	var auths []externalAuthRow
	if err := sentinel.Get("/api/core/entity/external/DISCORD", &auths); err != nil {
		return fmt.Errorf("list discord external auths: %w", err)
	}
	if len(auths) == 0 {
		return nil
	}

	allBindings, err := GetAllRoleBindings()
	if err != nil {
		return fmt.Errorf("load bindings: %w", err)
	}
	bindingsByGroup := make(map[string][]model.GroupDiscordRoleBinding, len(allBindings))
	for _, b := range allBindings {
		bindingsByGroup[b.GroupID] = append(bindingsByGroup[b.GroupID], b)
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
		if err := reconcileOneWithSnapshot(a.EntityID, a.ExternalID, bindingsByGroup, discordEnabled); err != nil {
			logger.SugarLogger.Errorf("group sync: reconcile failed for entity=%s discord=%s: %v", a.EntityID, a.ExternalID, err)
		}
	}
	logger.SugarLogger.Infof("group sync: full sweep complete")
	return nil
}

// reconcileOneWithSnapshot reconciles a single onboarded user using
// pre-fetched binding + allowed_sources snapshots. Reads current Discord
// roles via the guild-member lookup (state cache then REST fallback).
// Users no longer present in the guild are skipped rather than stripped —
// a transient lookup failure shouldn't aggressively remove memberships.
func reconcileOneWithSnapshot(entityID, discordID string, bindingsByGroup map[string][]model.GroupDiscordRoleBinding, discordEnabled map[string]bool) error {
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

	current, err := getEntityDiscordMemberships(entityID)
	if err != nil {
		return fmt.Errorf("fetch current memberships: %w", err)
	}
	currentSet := make(map[string]struct{}, len(current))
	for _, m := range current {
		currentSet[m.GroupID] = struct{}{}
	}

	for groupID := range desired {
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
