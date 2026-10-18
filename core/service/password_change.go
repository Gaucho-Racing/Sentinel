package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrPasswordUnchanged = errors.New("your new password must be different from your current one")
	ErrEmailNotVerified  = errors.New("verify your email before changing your password")
)

type InvalidPasswordError struct {
	error
}

func validateNewPassword(auth model.EntityEmail, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return InvalidPasswordError{err}
	}
	if bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(newPassword)) == nil {
		return ErrPasswordUnchanged
	}
	return nil
}

func verifiedEmailAuth(entityID string) (model.EntityEmail, error) {
	auth, err := GetEmailAuthForEntity(entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.EntityEmail{}, ErrNoEmailOnAccount
		}
		return model.EntityEmail{}, err
	}
	if auth.VerifiedAt == nil {
		return model.EntityEmail{}, ErrEmailNotVerified
	}
	return auth, nil
}

func SendPasswordChangeCode(ctx context.Context, entityID string, newPassword string) (EmailVerificationSent, error) {
	auth, err := verifiedEmailAuth(entityID)
	if err != nil {
		return EmailVerificationSent{}, err
	}
	if err := validateNewPassword(auth, newPassword); err != nil {
		return EmailVerificationSent{}, err
	}
	minutes := int(EmailVerificationCodeTTL.Minutes())
	return sendEmailCode(ctx, entityID, codePurposeChangePassword, auth.Email, func(code string) emailContent {
		return emailContent{
			Category:  "password-change",
			Subject:   "[Gaucho Racing] Confirm your password change",
			Heading:   "Confirm your password change",
			Intro:     "Enter this code in Sentinel to finish changing the password on your Gaucho Racing account.",
			Note:      "Never share this code with anyone. If you didn't request this, someone may be signed in to your account; contact team@gauchoracing.com.",
			Preheader: fmt.Sprintf("Your code is %s. It expires in %d minutes.", code, minutes),
			Body: fmt.Sprintf("Your Gaucho Racing password change code is %s.\n\n"+
				"Enter it in Sentinel to finish changing your password. It expires in %d minutes.\n\n"+
				"If you didn't request this, someone may be signed in to your account. Contact team@gauchoracing.com.", code, minutes),
		}
	})
}

func ChangePassword(ctx context.Context, entityID string, code string, newPassword string) error {
	auth, err := verifiedEmailAuth(entityID)
	if err != nil {
		return err
	}
	if err := validateNewPassword(auth, newPassword); err != nil {
		return err
	}
	verificationID, err := checkCode(entityID, codePurposeChangePassword, auth.Email, code)
	if err != nil {
		return err
	}
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := consumeCode(tx, verificationID); err != nil {
			return err
		}
		return tx.Model(&model.EntityEmail{}).Where("entity_id = ?", entityID).Update("password", hashed).Error
	})
	if err != nil {
		return err
	}
	logger.SugarLogger.Infof("Changed password for entity %s", entityID)

	changedAt := time.Now().UTC().Format("January 2, 2006 at 3:04 PM MST")
	if err := sendEmail(ctx, entityID, "", emailContent{
		Category:  "password-change",
		Subject:   "[Gaucho Racing] Your password was changed",
		Heading:   "Your password was changed",
		Intro:     "The password on your Gaucho Racing account was changed on " + changedAt + ".",
		Note:      "If this wasn't you, contact team@gauchoracing.com right away.",
		Preheader: "The password on your Gaucho Racing account was just changed.",
		Body: "The password on your Gaucho Racing account was changed on " + changedAt + ".\n\n" +
			"If this wasn't you, contact team@gauchoracing.com right away.",
	}); err != nil {
		logger.SugarLogger.Errorf("Failed to send password change notice to entity %s: %v", entityID, err)
	}
	return nil
}
