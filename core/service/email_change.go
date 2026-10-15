package service

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
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
		if err := recordEmailChange(tx, entityID, auth.Email, newEmail); err != nil {
			return err
		}
		return tx.Model(&model.EmailVerification{}).
			Where("entity_id = ? AND consumed_at IS NULL", entityID).
			Update("consumed_at", now).Error
	})
	if err != nil {
		return model.EntityEmail{}, err
	}
	logger.SugarLogger.Infof("Changed email for entity %s", entityID)
	TriggerProvisioningSync("email changed")
	auth.Email = newEmail
	auth.VerifiedAt = nil
	return auth, nil
}

func recordEmailChange(tx *gorm.DB, entityID string, oldEmail string, newEmail string) error {
	if oldEmail == "" || strings.EqualFold(oldEmail, newEmail) {
		return nil
	}
	return tx.Create(&model.EmailChange{
		ID:       ulid.Make().Prefixed("emc"),
		EntityID: entityID,
		OldEmail: oldEmail,
		NewEmail: newEmail,
	}).Error
}

// ListRetiredEmails returns addresses that belonged to an account before an
// email change and that no account uses now, lowercased.
func ListRetiredEmails() ([]string, error) {
	emails := []string{}
	err := database.DB.Raw(`
		SELECT DISTINCT LOWER(c.old_email)
		FROM auth_email_change c
		WHERE NOT EXISTS (
			SELECT 1 FROM auth_entity_email e WHERE LOWER(e.email) = LOWER(c.old_email)
		)
		ORDER BY 1
	`).Scan(&emails).Error
	return emails, err
}
