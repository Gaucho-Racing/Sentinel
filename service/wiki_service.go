package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
)

func GetAllWikiUsers() ([]model.WikiUser, error) {
	var userResponse model.WikiArrayResponse[model.WikiUser]
	var users []model.WikiUser

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://wiki.gauchoracing.com/api/users", nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return users, err
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return users, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return users, fmt.Errorf("failed to get all wiki users: %s", string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return users, err
	}
	users = userResponse.Data

	return users, nil
}

func GetWikiUserByID(id int) (model.WikiUser, error) {
	var user model.WikiUser

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return user, err
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return user, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return user, fmt.Errorf("failed to get wiki user: %s", string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return user, err
	}
	return user, nil
}

func CreateWikiUser(input model.WikiUserCreate) (int, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return 0, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://wiki.gauchoracing.com/api/users", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return 0, err
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to create wiki user: %s", string(body))
	}

	var user model.WikiUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return 0, err
	}
	utils.SugarLogger.Infof("Created wiki user: %s (%s)", user.Name, user.Email)
	return user.ID, nil
}

func UpdateWikiUser(id int, input model.WikiUserCreate) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), bytes.NewBuffer(jsonData))
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update wiki user: %s", string(body))
	}
	return nil
}

func DeleteWikiUser(id int) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://wiki.gauchoracing.com/api/users/%d", id), nil)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	req.Header.Set("Authorization", "Token "+config.WikiToken)

	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete wiki user: %s", string(body))
	}
	return nil
}

func CreateWikiUserWithPassword(password string, userID string) error {
	user := GetUserByID(userID)
	if user.ID == "" {
		return fmt.Errorf("user not found")
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
	id, err := CreateWikiUser(input)
	if err != nil {
		return err
	}
	er := GetRolesForUser(userID)
	er = append(er, "wiki_"+fmt.Sprint(id))
	SetRolesForUser(userID, er)
	return nil
}
