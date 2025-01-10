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

	// Check if scope contains "oidc" and add "user:read" if it does
	scopes := strings.Split(scope, " ")
	if contains(scopes, "oidc") && !contains(scopes, "user:read") {
		scopes = append(scopes, "user:read")
	}
	updatedScope := strings.Join(scopes, " ")

	authCode := model.AuthorizationCode{
		Code:      code,
		ClientID:  clientID,
		UserID:    userID,
		Scope:     updatedScope,
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

func GenerateIDToken(userID string, scope string, client_id string, expiresIn int) (string, error) {
	scopeList := strings.Split(scope, " ")
	filteredScopes := make([]string, 0)
	// only include openid scopes
	for _, s := range scopeList {
		if strings.HasPrefix(s, "openid") || strings.HasPrefix(s, "profile") || strings.HasPrefix(s, "email") || strings.HasPrefix(s, "roles") || strings.HasPrefix(s, "bookstack") {
			filteredScopes = append(filteredScopes, s)
		}
	}
	filteredScopes = append(filteredScopes, "user:read")
	filteredScope := strings.Join(filteredScopes, " ")
	return GenerateJWT(userID, filteredScope, client_id, expiresIn)
}

func SaveRefreshToken(token string, userID string, scope string, expiresIn int) error {
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Minute)
	refreshToken := model.RefreshToken{
		Token:     token,
		UserID:    userID,
		Scope:     scope,
		Revoked:   false,
		ExpiresAt: utils.WithPrecision(expiresAt),
	}
	result := database.DB.Create(&refreshToken)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func ValidateRefreshToken(token string) bool {
	var refreshToken model.RefreshToken
	database.DB.Where("token = ?", token).First(&refreshToken)
	if refreshToken.Token == "" {
		return false
	}
	if refreshToken.Revoked {
		return false
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		return false
	}
	return true
}

func RevokeRefreshToken(token string) error {
	database.DB.Model(&model.RefreshToken{}).Where("token = ?", token).Update("revoked", true)
	return nil
}
