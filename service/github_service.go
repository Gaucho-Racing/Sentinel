package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sentinel/config"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
	"strings"
)

func GetAllGithubUsers() ([]*model.GithubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/orgs/gaucho-racing/members", nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.GithubToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get all GitHub users: %s", string(body))
	}
	var githubUsers []*model.GithubUser
	err = json.NewDecoder(resp.Body).Decode(&githubUsers)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	return githubUsers, nil
}

func GetGithubStatusForUser(userID string) (*model.GithubOrgUser, error) {
	username := getGithubUsernameForUser(userID)
	if username == "" {
		return nil, fmt.Errorf("user does not have a GitHub account linked")
	}
	req, err := http.NewRequest("GET", "https://api.github.com/orgs/gaucho-racing/memberships/"+username, nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.GithubToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user membership status from GitHub organization: %s", string(body))
	}
	var githubUser *model.GithubOrgUser
	err = json.NewDecoder(resp.Body).Decode(&githubUser)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	return githubUser, nil
}

func AddUserToGithub(userID string, username string) error {
	user := GetUserByID(userID)
	reqBody := `{"role": "member"}`
	if user.IsInnerCircle() {
		reqBody = `{"role": "admin"}`
	}
	req, err := http.NewRequest("PUT", "https://api.github.com/orgs/gaucho-racing/memberships/"+username, strings.NewReader(reqBody))
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.GithubToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add user to GitHub organization: %s", string(body))
	}
	addGithubUsernameToRoles(username, userID)
	SendMessage(config.DiscordLogChannel, fmt.Sprintf("Added %s (%s) to GitHub organization: %s", username, user.Email, reqBody))
	return nil
}

func RemoveUserFromGithub(userID string, username string) error {
	req, err := http.NewRequest("DELETE", "https://api.github.com/orgs/gaucho-racing/memberships/"+username, nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.GithubToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return nil
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove user from GitHub organization: %s", string(body))
	}
	removeGithubUsernameFromRoles(username, userID)
	SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removed %s from GitHub organization", username))
	return nil
}

func PopulateGithubMembers() {
	users := GetAllUsers()
	for _, user := range users {
		ghUser := getGithubUsernameForUser(user.ID)
		if ghUser != "" {
			utils.SugarLogger.Infof("User %s has github username %s", user.ID, ghUser)
			AddUserToGithub(user.ID, ghUser)
		}
	}
}

func CleanGithubMembers() {
	keepUsers := []string{
		"gauchoracing",
	}
	githubUsers, err := GetAllGithubUsers()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	for _, ghUser := range githubUsers {
		user := getUserForGithubUsername(ghUser.Login)
		if user.ID == "" && !contains(keepUsers, ghUser.Login) {
			utils.SugarLogger.Infof("Removing user %s from GitHub organization", ghUser.Login)
			RemoveUserFromGithub(user.ID, ghUser.Login)
		} else if user.IsInnerCircle() {
			// keep inner circle members for now, in the future update perms appropriately
			orgUser, err := GetGithubStatusForUser(user.ID)
			if err != nil {
				utils.SugarLogger.Errorf("Error getting GitHub status for user %s: %s", user.ID, err.Error())
				continue
			}
			if orgUser.Role == "member" {
				AddUserToGithub(user.ID, ghUser.Login)
			}
		} else if user.IsMember() || user.IsAlumni() {
			orgUser, err := GetGithubStatusForUser(user.ID)
			if err != nil {
				utils.SugarLogger.Errorf("Error getting GitHub status for user %s: %s", user.ID, err.Error())
				continue
			}
			if orgUser.Role == "admin" {
				AddUserToGithub(user.ID, ghUser.Login)
			}
		} else {
			// User is not longer a member, remove from GitHub organization
			utils.SugarLogger.Infof("Removing user %s from GitHub organization", ghUser.Login)
			RemoveUserFromGithub(user.ID, ghUser.Login)
		}
	}
}

func addGithubUsernameToRoles(ghUsername string, userID string) {
	roles := GetRolesForUser(userID)
	for _, role := range roles {
		if strings.HasPrefix(role, "github_") {
			roles = removeValue(roles, role)
		}
	}
	roles = append(roles, "github_"+ghUsername)
	SetRolesForUser(userID, roles)
}

func removeGithubUsernameFromRoles(ghUsername string, userID string) {
	roles := GetRolesForUser(userID)
	newRoles := removeValue(roles, "github_"+ghUsername)
	SetRolesForUser(userID, newRoles)
}

func getGithubUsernameForUser(userID string) string {
	roles := GetRolesForUser(userID)
	for _, role := range roles {
		if strings.HasPrefix(role, "github_") {
			return strings.TrimPrefix(role, "github_")
		}
	}
	return ""
}

func getUserForGithubUsername(ghUsername string) model.User {
	var userID string
	database.DB.Table("user_role").Where("role = ?", "github_"+ghUsername).Select("user_id").Scan(&userID)
	if userID == "" {
		return model.User{}
	}
	return GetUserByID(userID)
}
