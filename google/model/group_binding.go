package model

import "time"

// GroupGoogleBinding maps a Sentinel group to the Google Group its membership
// is mirrored into. The relationship is 1:1 — one Sentinel group projects onto
// one Google Group, and a given Google Group is driven by a single Sentinel
// group — so both columns are unique.
//
// Owned by the google service: bindings are an integration-side concept that
// reference Sentinel group IDs from core but live in google's domain. Sync is
// one-way (Sentinel -> Google); this row only records where to project.
type GroupGoogleBinding struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	GroupID          string    `json:"group_id" gorm:"uniqueIndex"`
	GoogleGroupEmail string    `json:"google_group_email" gorm:"uniqueIndex"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (GroupGoogleBinding) TableName() string {
	return "group_google_binding"
}
