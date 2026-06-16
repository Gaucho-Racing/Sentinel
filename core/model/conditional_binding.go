package model

import "time"

// GroupConditionalBinding ties a Sentinel group to a set of *other* group
// IDs. A user matches the binding only if they're a member of EVERY group
// in RequiredGroupIDs (AND within a binding). Membership in the parent
// group is the OR across all bindings on the same parent (matches the
// Discord role-binding semantics).
//
// Membership is evaluated against ALL sources — a user counts as a member
// of a required group whether they got there via DIRECT, DISCORD, or
// another CONDITIONAL binding. That transitivity is what lets these
// bindings compose: "Senior CS" can be defined as "in CS Major AND in
// Class of 2026," where each of those may themselves be conditional.
//
// Cycles (A requires B, B requires A — or longer paths) are rejected at
// binding-creation time; see service.CheckBindingCycle.
type GroupConditionalBinding struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	GroupID          string      `json:"group_id" gorm:"index"`
	RequiredGroupIDs StringSlice `json:"required_group_ids" gorm:"type:jsonb"`
	CreatedAt        time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

func (GroupConditionalBinding) TableName() string {
	return "group_conditional_binding"
}
