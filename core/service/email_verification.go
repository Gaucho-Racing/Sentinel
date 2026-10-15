package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/gorm"
)

var ErrEmailAlreadyVerified = errors.New("email is already verified")

func SendEmailVerificationCode(ctx context.Context, entityID string) (EmailVerificationSent, error) {
	auth, err := GetEmailAuthForEntity(entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EmailVerificationSent{}, ErrNoEmailOnAccount
		}
		return EmailVerificationSent{}, err
	}
	if auth.VerifiedAt != nil {
		return EmailVerificationSent{}, ErrEmailAlreadyVerified
	}
	minutes := int(EmailVerificationCodeTTL.Minutes())
	return sendEmailCode(ctx, entityID, codePurposeVerifyEmail, auth.Email, func(code string) emailContent {
		return emailContent{
			Category:  "email-verification",
			Subject:   "[Gaucho Racing] Verify your email address",
			Heading:   "Confirm your email",
			Intro:     "Enter this code in Sentinel to verify the email on your Gaucho Racing account.",
			Note:      "Never share this code with anyone. If you didn't request this, you can safely ignore this email.",
			Preheader: fmt.Sprintf("Your verification code is %s. It expires in %d minutes.", code, minutes),
			Body: fmt.Sprintf("Your Gaucho Racing verification code is %s.\n\n"+
				"Enter it in Sentinel to verify the email on your account. It expires in %d minutes.\n\n"+
				"Never share this code with anyone. If you didn't request it, you can ignore this email.", code, minutes),
		}
	})
}

func ConfirmEmailVerificationCode(entityID string, code string) (model.EntityEmail, error) {
	auth, err := GetEmailAuthForEntity(entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.EntityEmail{}, ErrNoEmailOnAccount
		}
		return model.EntityEmail{}, err
	}
	if auth.VerifiedAt != nil {
		return auth, nil
	}
	verificationID, err := checkEmailCode(entityID, codePurposeVerifyEmail, auth.Email, code)
	if err != nil {
		return model.EntityEmail{}, err
	}

	now := time.Now()
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := consumeEmailCode(tx, verificationID); err != nil {
			return err
		}
		verified := tx.Model(&model.EntityEmail{}).
			Where("entity_id = ? AND email = ?", entityID, auth.Email).
			Update("verified_at", now)
		if verified.Error != nil {
			return verified.Error
		}
		if verified.RowsAffected == 0 {
			return ErrEmailVerificationNoCode
		}
		return nil
	})
	if err != nil {
		return model.EntityEmail{}, err
	}
	auth.VerifiedAt = &now
	logger.SugarLogger.Infof("Verified email for entity %s", entityID)
	return auth, nil
}
