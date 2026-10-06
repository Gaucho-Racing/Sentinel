package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

// ErrOrgPositionCycle is returned when a write would make a position its own
// ancestor. The API maps it to 400 — it's a bad request, not a server fault.
var ErrOrgPositionCycle = errors.New("position cannot be its own ancestor")

// ErrOrgPositionTitleRequired guards the one field with no sensible default.
var ErrOrgPositionTitleRequired = errors.New("title is required")

func GetAllOrgPositions() ([]model.OrgPosition, error) {
	positions := []model.OrgPosition{}
	if err := database.DB.Find(&positions).Error; err != nil {
		return []model.OrgPosition{}, err
	}
	sortOrgPositions(positions)
	return positions, nil
}

func GetOrgPositionByID(id string) (model.OrgPosition, error) {
	var position model.OrgPosition
	if err := database.DB.Where("id = ?", id).First(&position).Error; err != nil {
		return model.OrgPosition{}, err
	}
	return position, nil
}

func CreateOrgPosition(position model.OrgPosition) (model.OrgPosition, error) {
	if position.Title == "" {
		return model.OrgPosition{}, ErrOrgPositionTitleRequired
	}
	if position.ID == "" {
		position.ID = ulid.Make().Prefixed("orgp")
	}
	existing, err := GetAllOrgPositions()
	if err != nil {
		return model.OrgPosition{}, fmt.Errorf("load existing positions: %w", err)
	}
	if wouldCreateOrgCycle(position, existing) {
		return model.OrgPosition{}, ErrOrgPositionCycle
	}
	if err := database.DB.Create(&position).Error; err != nil {
		return model.OrgPosition{}, err
	}
	return position, nil
}

func UpdateOrgPosition(position model.OrgPosition) (model.OrgPosition, error) {
	if position.Title == "" {
		return model.OrgPosition{}, ErrOrgPositionTitleRequired
	}
	existing, err := GetAllOrgPositions()
	if err != nil {
		return model.OrgPosition{}, fmt.Errorf("load existing positions: %w", err)
	}
	if wouldCreateOrgCycle(position, existing) {
		return model.OrgPosition{}, ErrOrgPositionCycle
	}
	// Select the mutable columns explicitly so a zero value — clearing a holder
	// or detaching a group — actually writes, which Updates on a struct skips.
	if err := database.DB.Model(&model.OrgPosition{}).
		Where("id = ?", position.ID).
		Select("title", "parent_id", "entity_id", "group_id", "rank").
		Updates(position).Error; err != nil {
		return model.OrgPosition{}, err
	}
	return GetOrgPositionByID(position.ID)
}

// DeleteOrgPosition removes a position and lifts its children up to the
// deleted node's parent. Cascading instead would silently delete a whole
// subtree, and refusing would make reorganising tedious.
func DeleteOrgPosition(id string) error {
	position, err := GetOrgPositionByID(id)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.OrgPosition{}).
			Where("parent_id = ?", id).
			Update("parent_id", position.ParentID).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.OrgPosition{}).Error
	})
}

// sortOrgPositions puts siblings in rank order with a title tiebreak so the
// chart is stable across reloads regardless of insertion order.
func sortOrgPositions(positions []model.OrgPosition) {
	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].Rank != positions[j].Rank {
			return positions[i].Rank < positions[j].Rank
		}
		return positions[i].Title < positions[j].Title
	})
}

// wouldCreateOrgCycle walks up from the candidate's parent looking for the
// candidate itself. The visited set is defensive: if bad data ever put a cycle
// in the table, this still terminates instead of hanging the request.
func wouldCreateOrgCycle(candidate model.OrgPosition, existing []model.OrgPosition) bool {
	if candidate.ParentID == "" {
		return false
	}
	if candidate.ParentID == candidate.ID {
		return true
	}

	parentOf := make(map[string]string, len(existing))
	for _, p := range existing {
		parentOf[p.ID] = p.ParentID
	}
	// The candidate's own edge isn't in the table yet on create, and is stale
	// on update — overlay it so we test the post-write shape.
	parentOf[candidate.ID] = candidate.ParentID

	visited := make(map[string]bool)
	for node := candidate.ParentID; node != ""; node = parentOf[node] {
		if node == candidate.ID {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
	}
	return false
}
