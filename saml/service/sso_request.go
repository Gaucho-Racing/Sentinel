package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrAuthnRequestReplay = errors.New("SAML authentication request has already been used")

// GenerateSSORequest stashes a validated AuthnRequest for the SPA consent
// round-trip and returns its short-lived, single-use handle.
func GenerateSSORequest(requestID, spEntityID string, requestBuffer []byte, relayState string) (model.SSORequest, error) {
	if requestID == "" {
		return model.SSORequest{}, fmt.Errorf("SAML authentication request ID is required")
	}
	now := time.Now()
	if err := database.DB.Where("expires_at < ?", now).Delete(&model.SSORequest{}).Error; err != nil {
		return model.SSORequest{}, fmt.Errorf("clean expired SSO requests: %w", err)
	}
	req := model.SSORequest{
		ID:            generateCryptoString(32),
		RequestID:     requestID,
		SPEntityID:    spEntityID,
		RequestBuffer: string(requestBuffer),
		RelayState:    relayState,
		ExpiresAt:     now.Add(10 * time.Minute),
	}
	if err := database.DB.Create(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.SSORequest{}, ErrAuthnRequestReplay
		}
		return model.SSORequest{}, err
	}
	return req, nil
}

func ConsumeSSORequest(id string, consume func(model.SSORequest) error) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var req model.SSORequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&req).Error; err != nil {
			return fmt.Errorf("invalid sso request")
		}
		if time.Now().After(req.ExpiresAt) {
			if err := tx.Delete(&req).Error; err != nil {
				return err
			}
			return fmt.Errorf("sso request expired")
		}
		if err := consume(req); err != nil {
			return err
		}
		return tx.Delete(&req).Error
	})
}

// GetSSORequest loads a stashed request without consuming it. Used by the
// consent GET to look up the SP for the screen; the request is only consumed
// when the user approves.
func GetSSORequest(id string) (model.SSORequest, error) {
	var req model.SSORequest
	if err := database.DB.Where("id = ?", id).First(&req).Error; err != nil {
		return model.SSORequest{}, fmt.Errorf("invalid sso request")
	}
	if time.Now().After(req.ExpiresAt) {
		database.DB.Where("id = ?", id).Delete(&model.SSORequest{})
		return model.SSORequest{}, fmt.Errorf("sso request expired")
	}
	return req, nil
}

// DeleteSSORequest removes a stashed request. Called once the assertion has
// been issued so the handle is single-use, but only after success — a failed
// attempt leaves the stash in place so the user can retry without restarting
// the whole SP-initiated flow.
func generateCryptoString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
