package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
)

func GetAllWikiUsers() []model.WikiUser {
	var userResponse model.WikiArrayResponse[model.WikiUser]
	var users []model.WikiUser

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://wiki.gauchoracing.com/api/users", nil)
	if err != nil {
		utils.SugarLogger.Errorf("Error creating request: %v", err)
		return users
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorf("Error making request: %v", err)
		return users
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		utils.SugarLogger.Errorf("Unexpected status code: %d", resp.StatusCode)
		return users
	}

	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		utils.SugarLogger.Errorf("Error decoding response: %v", err)
		return users
	}
	users = userResponse.Data

	return users
}

func GetWikiUserByID(id int) model.WikiUser {
	var user model.WikiUser

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), nil)
	if err != nil {
		utils.SugarLogger.Errorf("Error creating request: %v", err)
		return user
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorf("Error making request: %v", err)
		return user
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		utils.SugarLogger.Errorf("Unexpected status code: %d", resp.StatusCode)
		return user
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		utils.SugarLogger.Errorf("Error decoding response: %v", err)
		return user
	}
	return user
}

func CreateWikiUser(input model.WikiUserCreate) int {
	jsonData, err := json.Marshal(input)
	if err != nil {
		utils.SugarLogger.Errorf("Error marshaling input: %v", err)
		return 0
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://wiki.gauchoracing.com/api/users", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.SugarLogger.Errorf("Error creating request: %v", err)
		return 0
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorf("Error making request: %v", err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		utils.SugarLogger.Errorf("Unexpected status code: %d", resp.StatusCode)
		return 0
	}

	var user model.WikiUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		utils.SugarLogger.Errorf("Error decoding response: %v", err)
		return 0
	}
	utils.SugarLogger.Infof("Created wiki user: %s (%s)", user.Name, user.Email)
	return user.ID
}

func UpdateWikiUser(id int, input model.WikiUserCreate) bool {
	jsonData, err := json.Marshal(input)
	if err != nil {
		utils.SugarLogger.Errorf("Error marshaling input: %v", err)
		return false
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), bytes.NewBuffer(jsonData))
	if err != nil {
		utils.SugarLogger.Errorf("Error creating request: %v", err)
		return false
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorf("Error making request: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		utils.SugarLogger.Errorf("Unexpected status code: %d", resp.StatusCode)
		return false
	}

	return true
}

func DeleteWikiUser(id int) bool {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), nil)
	if err != nil {
		utils.SugarLogger.Errorf("Error creating request: %v", err)
		return false
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorf("Error making request: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		utils.SugarLogger.Errorf("Unexpected status code: %d", resp.StatusCode)
		return false
	}

	return true
}

func CreateWikiUserWithPassword(password string, userID string) int {
	user := GetUserByID(userID)
	if user.ID == "" {
		return 0
	}
	roles := []int{int(model.WikiRoleEditor)}
	if user.IsInnerCircle() {
		roles = append(roles, int(model.WikiRoleLead))
	}
	input := model.WikiUserCreate{
		Name:           user.FirstName + " " + user.LastName,
		Email:          user.Email,
		Roles:          roles,
		ExternalAuthID: userID,
		Password:       password,
		SendInvite:     false,
	}
	id := CreateWikiUser(input)
	if id == 0 {
		return 0
	}
	er := GetRolesForUser(userID)
	er = append(er, "wiki_"+fmt.Sprint(id))
	SetRolesForUser(userID, er)
	return id
}
