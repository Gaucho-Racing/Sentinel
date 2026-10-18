package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/pkg/notify"
	"gorm.io/gorm"
)

const codePurposeVerifyPhone = "verify_phone"

var e164 = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

var (
	ErrNoPhoneOnAccount     = errors.New("account has no phone number")
	ErrPhoneAlreadyVerified = errors.New("phone number is already verified")
	ErrInvalidPhone         = errors.New("enter a valid phone number, including the country code")
	ErrPhoneUnchanged       = errors.New("that's already your phone number")
	ErrPhoneInUse           = errors.New("that phone number is already used by another account")
)

type PhoneVerificationSent struct {
	PhoneNumber string    `json:"phone_number"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

// NormalizePhoneNumber strips common formatting and accepts only E.164.
func NormalizePhoneNumber(value string) (string, bool) {
	normalized := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '-', '(', ')', '.':
			return -1
		}
		return r
	}, strings.TrimSpace(value))
	return normalized, e164.MatchString(normalized)
}

func phoneAuthForEntity(entityID string) (model.EntityPhone, error) {
	auth, err := GetPhoneAuthForEntity(entityID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && auth.PhoneNumber == "") {
		return model.EntityPhone{}, ErrNoPhoneOnAccount
	}
	return auth, err
}

func SendPhoneVerificationCode(ctx context.Context, entityID string) (PhoneVerificationSent, error) {
	auth, err := phoneAuthForEntity(entityID)
	if err != nil {
		return PhoneVerificationSent{}, err
	}
	if auth.VerifiedAt != nil {
		return PhoneVerificationSent{}, ErrPhoneAlreadyVerified
	}
	verification, err := issueCode(ctx, entityID, codePurposeVerifyPhone, auth.PhoneNumber, func(ctx context.Context, code string) error {
		return sendSMS(ctx, entityID, fmt.Sprintf(
			"Gaucho Racing: Your verification code is %s. It expires in %d minutes. Never share this code.",
			code, int(EmailVerificationCodeTTL.Minutes()),
		))
	})
	if err != nil {
		return PhoneVerificationSent{}, err
	}
	return PhoneVerificationSent{
		PhoneNumber: auth.PhoneNumber,
		ExpiresAt:   verification.ExpiresAt,
		ResendAfter: verification.CreatedAt.Add(EmailVerificationResendAfter),
	}, nil
}

func ConfirmPhoneVerificationCode(entityID string, code string) (model.EntityPhone, error) {
	auth, err := phoneAuthForEntity(entityID)
	if err != nil {
		return model.EntityPhone{}, err
	}
	if auth.VerifiedAt != nil {
		return auth, nil
	}
	verificationID, err := checkCode(entityID, codePurposeVerifyPhone, auth.PhoneNumber, code)
	if err != nil {
		return model.EntityPhone{}, err
	}
	now := time.Now()
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := consumeCode(tx, verificationID); err != nil {
			return err
		}
		verified := tx.Model(&model.EntityPhone{}).
			Where("entity_id = ? AND phone_number = ?", entityID, auth.PhoneNumber).
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
		return model.EntityPhone{}, err
	}
	auth.VerifiedAt = &now
	logger.SugarLogger.Infof("Verified phone for entity %s", entityID)
	return auth, nil
}

// ChangePhone replaces the entity's phone number with an unverified one.
func ChangePhone(entityID string, phoneNumber string) (model.EntityPhone, error) {
	normalized, ok := NormalizePhoneNumber(phoneNumber)
	if !ok {
		return model.EntityPhone{}, ErrInvalidPhone
	}
	current, err := GetPhoneAuthForEntity(entityID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.EntityPhone{}, err
	}
	if current.PhoneNumber == normalized {
		return model.EntityPhone{}, ErrPhoneUnchanged
	}
	var taken int64
	if err := database.DB.Model(&model.EntityPhone{}).
		Where("phone_number = ? AND entity_id <> ?", normalized, entityID).
		Count(&taken).Error; err != nil {
		return model.EntityPhone{}, err
	}
	if taken > 0 {
		return model.EntityPhone{}, ErrPhoneInUse
	}

	updated := model.EntityPhone{EntityID: entityID, PhoneNumber: normalized}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if current.EntityID == "" {
			if err := tx.Create(&updated).Error; err != nil {
				return err
			}
		} else if err := tx.Model(&model.EntityPhone{}).
			Where("entity_id = ?", entityID).
			Updates(map[string]any{"phone_number": normalized, "verified_at": nil}).Error; err != nil {
			return err
		}
		return tx.Model(&model.EmailVerification{}).
			Where("entity_id = ? AND purpose = ? AND consumed_at IS NULL", entityID, codePurposeVerifyPhone).
			Update("consumed_at", time.Now()).Error
	})
	if err != nil {
		return model.EntityPhone{}, err
	}
	logger.SugarLogger.Infof("Changed phone number for entity %s", entityID)
	return updated, nil
}

func sendSMS(ctx context.Context, entityID string, body string) error {
	token, err := CoreServiceToken()
	if err != nil {
		return fmt.Errorf("mint core service token: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, emailVerificationNotifyTimeout)
	defer cancel()
	client := notify.Client{BaseURL: config.NotifyURL, Token: token}
	return client.SendSMS(ctx, notify.SMS{
		EntityID: entityID,
		SenderID: config.NotifySMSSenderID,
		Category: "phone-verification",
		Body:     body,
	})
}
