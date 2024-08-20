package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
	"strings"
)

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
		return nil, fmt.Errorf("failed to get user membership status from GitHub organization")
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
	return nil
}

func addGithubUsernameToRoles(ghUsername string, userID string) {
	roles := GetRolesForUser(userID)
	roles = append(roles, "github_"+ghUsername)
	SetRolesForUser(userID, roles)
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
