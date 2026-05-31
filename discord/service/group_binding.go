package service

import (
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllRoleBindings() ([]model.GroupDiscordRoleBinding, error) {
	bindings := []model.GroupDiscordRoleBinding{}
	if err := database.DB.Find(&bindings).Error; err != nil {
		return []model.GroupDiscordRoleBinding{}, err
	}
	return bindings, nil
}

func GetRoleBindingsForGroup(groupID string) ([]model.GroupDiscordRoleBinding, error) {
	bindings := []model.GroupDiscordRoleBinding{}
	if err := database.DB.Where("group_id = ?", groupID).Find(&bindings).Error; err != nil {
		return []model.GroupDiscordRoleBinding{}, err
	}
	return bindings, nil
}

func CreateRoleBinding(binding model.GroupDiscordRoleBinding) (model.GroupDiscordRoleBinding, error) {
	if binding.ID == "" {
		binding.ID = ulid.Make().Prefixed("gdrb")
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		return model.GroupDiscordRoleBinding{}, err
	}
	return binding, nil
}

// DeleteRoleBinding scopes the delete to (groupID, bindingID) so a tampered
// request can't drop a binding from a different group.
func DeleteRoleBinding(groupID, bindingID string) error {
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, bindingID).Delete(&model.GroupDiscordRoleBinding{}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteAllRoleBindingsForGroup(groupID string) error {
	if err := database.DB.Where("group_id = ?", groupID).Delete(&model.GroupDiscordRoleBinding{}).Error; err != nil {
		return err
	}
	return nil
}

// EvaluateDiscordMembership reports whether a user holding userRoles
// satisfies any of the given Discord role bindings. Within a binding, all
// listed roles must be held (AND); across bindings, any single match
// qualifies the user (OR). A binding with no roles never matches — an
// empty AND-group is treated as a no-match rather than a vacuous grant.
func EvaluateDiscordMembership(bindings []model.GroupDiscordRoleBinding, userRoles []string) bool {
	if len(bindings) == 0 {
		return false
	}
	held := make(map[string]struct{}, len(userRoles))
	for _, r := range userRoles {
		held[r] = struct{}{}
	}
	for _, b := range bindings {
		if len(b.DiscordRoleIDs) == 0 {
			continue
		}
		matched := true
		for _, required := range b.DiscordRoleIDs {
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

// GetEligibleGroupsForUserRoles returns the IDs of all groups whose Discord
// role bindings are satisfied by userRoles. Groups with no bindings are not
// included.
func GetEligibleGroupsForUserRoles(userRoles []string) ([]string, error) {
	all, err := GetAllRoleBindings()
	if err != nil {
		return nil, err
	}
	byGroup := make(map[string][]model.GroupDiscordRoleBinding)
	for _, b := range all {
		byGroup[b.GroupID] = append(byGroup[b.GroupID], b)
	}
	eligible := make([]string, 0, len(byGroup))
	for groupID, bs := range byGroup {
		if EvaluateDiscordMembership(bs, userRoles) {
			eligible = append(eligible, groupID)
		}
	}
	return eligible, nil
}
