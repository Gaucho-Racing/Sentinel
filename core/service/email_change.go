package service

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail      = errors.New("enter a valid email address")
	ErrEmailUnchanged    = errors.New("that's already your email address")
	ErrEmailInUse        = errors.New("that email address is already used by another account")
	ErrIncorrectPassword = errors.New("incorrect password")
)

func ChangeEmail(entityID string, newEmail string) (model.EntityEmail, error) {
	newEmail = strings.TrimSpace(newEmail)
	parsed, err := mail.ParseAddress(newEmail)
	if err != nil || parsed.Address != newEmail {
		return model.EntityEmail{}, ErrInvalidEmail
	}

	auth, err := GetEmailAuthForEntity(entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.EntityEmail{}, ErrNoEmailOnAccount
		}
		return model.EntityEmail{}, err
	}
	if strings.EqualFold(auth.Email, newEmail) {
		return model.EntityEmail{}, ErrEmailUnchanged
	}

	var taken int64
	if err := database.DB.Model(&model.EntityEmail{}).
		Where("LOWER(email) = LOWER(?) AND entity_id <> ?", newEmail, entityID).
		Count(&taken).Error; err != nil {
		return model.EntityEmail{}, err
	}
	if taken > 0 {
		return model.EntityEmail{}, ErrEmailInUse
	}

	now := time.Now()
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.EntityEmail{}).
			Where("entity_id = ?", entityID).
			Updates(map[string]any{"email": newEmail, "verified_at": nil}).Error; err != nil {
			return err
		}
		return tx.Model(&model.EmailVerification{}).
			Where("entity_id = ? AND consumed_at IS NULL", entityID).
			Update("consumed_at", now).Error
	})
	if err != nil {
		return model.EntityEmail{}, err
	}
	logger.SugarLogger.Infof("Changed email for entity %s from %s to %s", entityID, auth.Email, newEmail)
	auth.Email = newEmail
	auth.VerifiedAt = nil
	return auth, nil
}
