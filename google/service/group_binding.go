package service

import (
	"errors"

	"github.com/gaucho-racing/sentinel/google/database"
	"github.com/gaucho-racing/sentinel/google/model"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

var ErrBindingNotFound = errors.New("group google binding not found")

func GetAllGoogleBindings() ([]model.GroupGoogleBinding, error) {
	bindings := []model.GroupGoogleBinding{}
	if err := database.DB.Find(&bindings).Error; err != nil {
		return []model.GroupGoogleBinding{}, err
	}
	return bindings, nil
}

// GetGoogleBindingForGroup returns the binding for a group, or
// ErrBindingNotFound if the group has none. The 1:1 model means at most one
// row per group.
func GetGoogleBindingForGroup(groupID string) (model.GroupGoogleBinding, error) {
	var binding model.GroupGoogleBinding
	if err := database.DB.Where("group_id = ?", groupID).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GroupGoogleBinding{}, ErrBindingNotFound
		}
		return model.GroupGoogleBinding{}, err
	}
	return binding, nil
}

func CreateGoogleBinding(binding model.GroupGoogleBinding) (model.GroupGoogleBinding, error) {
	if binding.ID == "" {
		binding.ID = ulid.Make().Prefixed("ggb")
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		return model.GroupGoogleBinding{}, err
	}
	return binding, nil
}

// DeleteGoogleBinding scopes the delete to (groupID, bindingID) so a tampered
// request can't drop a binding for a different group.
func DeleteGoogleBinding(groupID, bindingID string) error {
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, bindingID).Delete(&model.GroupGoogleBinding{}).Error; err != nil {
		return err
	}
	return nil
}
