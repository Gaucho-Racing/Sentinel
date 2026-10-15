package service

import (
	"errors"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	ReauthScope    = "sentinel:reauth"
	ReauthTokenTTL = 5 * time.Minute
)

var ErrReauthRequired = errors.New("confirm your password to continue")

type ReauthGrant struct {
	Token     string    `json:"reauth_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func Reauthenticate(entityID string, clientID string, password string) (ReauthGrant, error) {
	auth, err := GetEmailAuthForEntity(entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ReauthGrant{}, ErrNoEmailOnAccount
		}
		return ReauthGrant{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password)) != nil {
		return ReauthGrant{}, ErrIncorrectPassword
	}
	expiresAt := time.Now().Add(ReauthTokenTTL)
	token, _, err := GenerateToken(entityID, clientID, ReauthScope, int(ReauthTokenTTL.Seconds()), nil)
	if err != nil {
		return ReauthGrant{}, err
	}
	return ReauthGrant{Token: token, ExpiresAt: expiresAt}, nil
}

func CheckReauthToken(entityID string, token string) (string, error) {
	if token == "" {
		return "", ErrReauthRequired
	}
	claims, err := ValidateToken(token)
	if err != nil || claims.Subject != entityID || claims.Scope != ReauthScope {
		return "", ErrReauthRequired
	}
	return claims.ID, nil
}

func RevokeReauthToken(tokenID string) {
	if err := database.DB.Delete(&model.Token{}, "id = ?", tokenID).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to revoke reauth token %s: %v", tokenID, err)
	}
}
