package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/oauth/database"
	"github.com/gaucho-racing/sentinel/oauth/model"
)

func GenerateAuthorizationCode(entityID string, clientID string, scope string, redirectURI string) (model.AuthorizationCode, error) {
	code := generateCryptoString(32)
	authCode := model.AuthorizationCode{
		Code:        code,
		EntityID:    entityID,
		ClientID:    clientID,
		Scope:       scope,
		RedirectURI: redirectURI,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}
	if err := database.DB.Create(&authCode).Error; err != nil {
		return model.AuthorizationCode{}, err
	}
	return authCode, nil
}

func VerifyAuthorizationCode(code string) (model.AuthorizationCode, error) {
	var authCode model.AuthorizationCode
	if err := database.DB.Where("code = ?", code).First(&authCode).Error; err != nil {
		return model.AuthorizationCode{}, fmt.Errorf("invalid authorization code")
	}
	defer database.DB.Where("code = ?", code).Delete(&model.AuthorizationCode{})
	if time.Now().After(authCode.ExpiresAt) {
		return model.AuthorizationCode{}, fmt.Errorf("authorization code expired")
	}
	return authCode, nil
}

func generateCryptoString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
