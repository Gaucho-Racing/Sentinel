package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
)

// GenerateSSORequest stashes a validated AuthnRequest for the SPA consent
// round-trip and returns its short-lived, single-use handle.
func GenerateSSORequest(spEntityID string, requestBuffer []byte, relayState string) (model.SSORequest, error) {
	req := model.SSORequest{
		ID:            generateCryptoString(32),
		SPEntityID:    spEntityID,
		RequestBuffer: string(requestBuffer),
		RelayState:    relayState,
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
	if err := database.DB.Create(&req).Error; err != nil {
		return model.SSORequest{}, err
	}
	return req, nil
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
func DeleteSSORequest(id string) {
	database.DB.Where("id = ?", id).Delete(&model.SSORequest{})
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
