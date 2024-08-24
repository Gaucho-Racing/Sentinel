package service

import (
	"crypto/rand"
	"fmt"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
	"strings"
	"time"
)

func GetAllClientApplications() []model.ClientApplication {
	var clientApplications []model.ClientApplication
	database.DB.Order("name asc").Find(&clientApplications)
	for i := range clientApplications {
		clientApplications[i].RedirectURIs = GetRedirectURIsForClientApplication(clientApplications[i].ID)
	}
	return clientApplications
}

func GetClientApplicationsForUser(userID string) []model.ClientApplication {
	var clientApplications []model.ClientApplication
	database.DB.Where("user_id = ?", userID).Order("name asc").Find(&clientApplications)
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
		clientApplication.ID = generateCryptoString(12)
		clientApplication.Secret = generateCryptoString(32)
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

func generateCryptoString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func ValidateRedirectURI(uri string, clientID string) bool {
	validUris := GetRedirectURIsForClientApplication(clientID)
	return contains(validUris, uri)
}

func ValidateScope(scopes string) bool {
	validScopes := []string{}
	for k := range model.ValidOauthScopes {
		validScopes = append(validScopes, k)
	}
	inputScopes := strings.Split(scopes, " ")
	for _, scope := range inputScopes {
		if !contains(validScopes, scope) {
			return false
		}
	}
	return true
}

func GenerateAuthorizationCode(clientID, userID, scope string) (model.AuthorizationCode, error) {
	code := generateCryptoString(8)
	expiresAt := time.Now().Add(5 * time.Minute)
	authCode := model.AuthorizationCode{
		Code:      code,
		ClientID:  clientID,
		UserID:    userID,
		Scope:     scope,
		ExpiresAt: utils.WithPrecision(expiresAt),
	}
	result := database.DB.Create(&authCode)
	if result.Error != nil {
		return authCode, result.Error
	}
	return authCode, nil
}

func VerifyAuthorizationCode(code string) (model.AuthorizationCode, error) {
	var authCode model.AuthorizationCode
	database.DB.Where("code = ?", code).First(&authCode)
	if authCode.Code == "" {
		return model.AuthorizationCode{}, fmt.Errorf("invalid code")
	}
	defer database.DB.Delete(&authCode)
	if time.Now().After(authCode.ExpiresAt) {
		return model.AuthorizationCode{}, fmt.Errorf("invalid code")
	}
	return authCode, nil
}
