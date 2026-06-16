package service

import (
	"errors"
	"fmt"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/ulid-go"
)

// ErrConditionalBindingCycle is returned when CreateConditionalBinding would
// close a cycle in the binding dependency graph (e.g. A requires B, B
// requires A). The API layer maps this to 400 so the admin gets a clear
// rejection rather than a server error.
var ErrConditionalBindingCycle = errors.New("conditional binding would create a cycle")

// ErrConditionalBindingSelfRef is returned when a binding lists its own
// parent group in RequiredGroupIDs — the degenerate single-node cycle.
var ErrConditionalBindingSelfRef = errors.New("conditional binding cannot require its own group")

func GetAllConditionalBindings() ([]model.GroupConditionalBinding, error) {
	bindings := []model.GroupConditionalBinding{}
	if err := database.DB.Find(&bindings).Error; err != nil {
		return []model.GroupConditionalBinding{}, err
	}
	return bindings, nil
}

func GetConditionalBindingsForGroup(groupID string) ([]model.GroupConditionalBinding, error) {
	bindings := []model.GroupConditionalBinding{}
	if err := database.DB.Where("group_id = ?", groupID).Find(&bindings).Error; err != nil {
		return []model.GroupConditionalBinding{}, err
	}
	return bindings, nil
}

// CreateConditionalBinding validates the binding against existing ones for
// cycles, mints an ID if absent, and inserts. Returns the created row.
func CreateConditionalBinding(binding model.GroupConditionalBinding) (model.GroupConditionalBinding, error) {
	// Self-reference is the trivial cycle — catch it explicitly so the error
	// is unambiguous (vs. surfacing as a generic "creates a cycle").
	for _, req := range binding.RequiredGroupIDs {
		if req == binding.GroupID {
			return model.GroupConditionalBinding{}, ErrConditionalBindingSelfRef
		}
	}

	existing, err := GetAllConditionalBindings()
	if err != nil {
		return model.GroupConditionalBinding{}, fmt.Errorf("load existing bindings: %w", err)
	}
	if wouldCreateCycle(binding, existing) {
		return model.GroupConditionalBinding{}, ErrConditionalBindingCycle
	}

	if binding.ID == "" {
		binding.ID = ulid.Make().Prefixed("gcb")
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		return model.GroupConditionalBinding{}, err
	}
	return binding, nil
}

// DeleteConditionalBinding scopes the delete to (groupID, bindingID) so a
// tampered request can't drop a binding from a different group.
func DeleteConditionalBinding(groupID, bindingID string) error {
	return database.DB.
		Where("group_id = ? AND id = ?", groupID, bindingID).
		Delete(&model.GroupConditionalBinding{}).Error
}

func DeleteAllConditionalBindingsForGroup(groupID string) error {
	return database.DB.
		Where("group_id = ?", groupID).
		Delete(&model.GroupConditionalBinding{}).Error
}

// EvaluateConditionalMembership reports whether an entity that's a member of
// entityGroupIDs satisfies any of the given conditional bindings. Within a
// binding, ALL required groups must be held (AND); across bindings on the
// same parent, ANY match qualifies the user (OR). A binding with no
// required groups never matches — an empty AND-group is a no-match rather
// than a vacuous grant. Mirrors EvaluateDiscordMembership exactly.
func EvaluateConditionalMembership(bindings []model.GroupConditionalBinding, entityGroupIDs []string) bool {
	if len(bindings) == 0 {
		return false
	}
	held := make(map[string]struct{}, len(entityGroupIDs))
	for _, g := range entityGroupIDs {
		held[g] = struct{}{}
	}
	for _, b := range bindings {
		if len(b.RequiredGroupIDs) == 0 {
			continue
		}
		matched := true
		for _, required := range b.RequiredGroupIDs {
			if _, ok := held[required]; !ok {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

// wouldCreateCycle returns true if adding `newBinding` to `existing` would
// close a cycle in the dependency graph (parent group → required groups).
// DFS from each of the new binding's required groups; if we can reach the
// new binding's own GroupID, that's a cycle.
func wouldCreateCycle(newBinding model.GroupConditionalBinding, existing []model.GroupConditionalBinding) bool {
	// Build adjacency: for each group, the union of required-group IDs across
	// all its bindings. (Multiple bindings on the same parent = OR semantics
	// at evaluation time, but for cycle detection any required group from any
	// binding creates a dependency edge.)
	adj := make(map[string]map[string]struct{})
	add := func(parent, req string) {
		if adj[parent] == nil {
			adj[parent] = make(map[string]struct{})
		}
		adj[parent][req] = struct{}{}
	}
	for _, b := range existing {
		for _, req := range b.RequiredGroupIDs {
			add(b.GroupID, req)
		}
	}
	for _, req := range newBinding.RequiredGroupIDs {
		add(newBinding.GroupID, req)
	}

	// DFS from each required group looking for a path back to the new
	// binding's parent. visited prevents infinite loops in case the EXISTING
	// graph already has cycles (which shouldn't happen if we always check on
	// create, but defensive).
	target := newBinding.GroupID
	visited := make(map[string]bool)
	var dfs func(node string) bool
	dfs = func(node string) bool {
		if node == target {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		for next := range adj[node] {
			if dfs(next) {
				return true
			}
		}
		return false
	}
	for _, req := range newBinding.RequiredGroupIDs {
		if dfs(req) {
			return true
		}
	}
	return false
}
