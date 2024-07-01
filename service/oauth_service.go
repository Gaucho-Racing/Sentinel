package service

import (
	"fmt"
	"math/rand"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
)

func GetAllClientApplications() []model.ClientApplication {
	var clientApplications []model.ClientApplication
	database.DB.Find(&clientApplications)
	for i := range clientApplications {
		clientApplications[i].RedirectURIs = GetRedirectURIsForClientApplication(clientApplications[i].ID)
	}
	return clientApplications
}

func GetClientApplicationsForUser(userID string) []model.ClientApplication {
	var clientApplications []model.ClientApplication
	database.DB.Where("user_id = ?", userID).Find(&clientApplications)
	for i := range clientApplications {
		clientApplications[i].RedirectURIs = GetRedirectURIsForClientApplication(clientApplications[i].ID)
	}
	return clientApplications
}

func GetClientApplicationByID(clientID string) model.ClientApplication {
	var clientApplication model.ClientApplication
	database.DB.Where("id = ?", clientID).First(&clientApplication)
	clientApplication.RedirectURIs = GetRedirectURIsForClientApplication(clientID)
	return clientApplication
}

func CreateClientApplication(clientApplication model.ClientApplication) (model.ClientApplication, error) {
	if clientApplication.ID == "" {
		clientApplication.ID = generateClientID()
		clientApplication.Secret = generateClientSecret()
	} else {
		existing := GetClientApplicationByID(clientApplication.ID)
		if existing.ID != "" {
			clientApplication.Secret = existing.Secret
		} else {
			return model.ClientApplication{}, fmt.Errorf("client application with id: %s does not exist", clientApplication.ID)
		}
	}
	if clientApplication.Name == "" {
		return model.ClientApplication{}, fmt.Errorf("client application name cannot be empty")
	}
	user := GetUserByID(clientApplication.UserID)
	if user.ID == "" {
		return model.ClientApplication{}, fmt.Errorf("user with id: %s does not exist", clientApplication.UserID)
	}
	if database.DB.Where("id = ?", clientApplication.ID).Updates(&clientApplication).RowsAffected == 0 {
		utils.SugarLogger.Infof("New client application created with id: %s", clientApplication.ID)
		if result := database.DB.Create(&clientApplication); result.Error != nil {
			return model.ClientApplication{}, result.Error
		}
	} else {
		utils.SugarLogger.Infof("Client application with id: %s has been updated!", clientApplication.ID)
	}
	SetRedirectURIsForClientApplication(clientApplication.ID, clientApplication.RedirectURIs)
	return GetClientApplicationByID(clientApplication.ID), nil
}

func DeleteClientApplication(clientID string) error {
	if result := database.DB.Where("id = ?", clientID).Delete(&model.ClientApplication{}); result.Error != nil {
		return result.Error
	}
	SetRedirectURIsForClientApplication(clientID, []string{})
	return nil
}

func GetRedirectURIsForClientApplication(clientID string) []string {
	var redirectURIs []model.ClientApplicationRedirectURI
	database.DB.Where("client_application_id = ?", clientID).Find(&redirectURIs)
	uriStrings := make([]string, len(redirectURIs))
	for i, uri := range redirectURIs {
		uriStrings[i] = uri.RedirectURI
	}
	return uriStrings
}

func SetRedirectURIsForClientApplication(clientID string, redirectURIs []string) []string {
	existingURIs := GetRedirectURIsForClientApplication(clientID)
	for _, nr := range redirectURIs {
		if !contains(existingURIs, nr) {
			database.DB.Create(&model.ClientApplicationRedirectURI{
				ClientApplicationID: clientID,
				RedirectURI:         nr,
			})
		}
	}
	for _, er := range existingURIs {
		if !contains(redirectURIs, er) {
			database.DB.Where("client_application_id = ? AND redirect_uri = ?", clientID, er).Delete(&model.ClientApplicationRedirectURI{})
		}
	}
	return GetRedirectURIsForClientApplication(clientID)
}

func generateClientID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 12

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func generateClientSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 32

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
