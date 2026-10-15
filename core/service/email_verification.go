package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/pkg/notify"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

const (
	EmailVerificationCodeTTL       = 10 * time.Minute
	EmailVerificationResendAfter   = 60 * time.Second
	EmailVerificationMaxAttempts   = 5
	emailVerificationMaxPerHour    = 5
	emailVerificationCodeDigits    = 6
	emailVerificationNotifyTimeout = 10 * time.Second
)

var (
	ErrEmailAlreadyVerified        = errors.New("email is already verified")
	ErrNoEmailOnAccount            = errors.New("account has no email address")
	ErrEmailVerificationNotSent    = errors.New("could not send the verification email")
	ErrEmailVerificationNoCode     = errors.New("no active verification code; request a new one")
	ErrEmailVerificationWrongCode  = errors.New("incorrect verification code")
	ErrEmailVerificationTooManyTry = errors.New("too many incorrect attempts; request a new code")
)

type EmailVerificationThrottled struct {
	RetryAfter time.Duration
}

func (e EmailVerificationThrottled) Error() string {
	return fmt.Sprintf("too many verification emails; try again in %d seconds", int(e.RetryAfter.Seconds()))
}

type EmailVerificationSent struct {
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

//go:embed templates/email_verification.html
var emailTemplates embed.FS

var emailVerificationTemplate = template.Must(template.ParseFS(emailTemplates, "templates/email_verification.html"))

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
	if err := checkEmailVerificationThrottle(entityID, auth.Email); err != nil {
		return EmailVerificationSent{}, err
	}

	code, err := generateVerificationCode()
	if err != nil {
		return EmailVerificationSent{}, err
	}
	now := time.Now()
	verification := model.EmailVerification{
		ID:        ulid.Make().Prefixed("emv"),
		EntityID:  entityID,
		Email:     auth.Email,
		ExpiresAt: now.Add(EmailVerificationCodeTTL),
	}
	verification.CodeHash = hashVerificationCode(verification.ID, code)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.EmailVerification{}).
			Where("entity_id = ? AND consumed_at IS NULL", entityID).
			Update("consumed_at", now).Error; err != nil {
			return err
		}
		return tx.Create(&verification).Error
	})
	if err != nil {
		return EmailVerificationSent{}, err
	}

	if err := sendVerificationEmail(ctx, entityID, code); err != nil {
		logger.SugarLogger.Errorf("Failed to send email verification code to entity %s: %v", entityID, err)
		database.DB.Delete(&model.EmailVerification{}, "id = ?", verification.ID)
		return EmailVerificationSent{}, ErrEmailVerificationNotSent
	}

	return EmailVerificationSent{
		Email:       auth.Email,
		ExpiresAt:   verification.ExpiresAt,
		ResendAfter: verification.CreatedAt.Add(EmailVerificationResendAfter),
	}, nil
}

func ConfirmEmailVerificationCode(entityID string, code string) (model.EntityEmail, error) {
	code = strings.TrimSpace(code)
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

	var verification model.EmailVerification
	err = database.DB.
		Where("entity_id = ? AND consumed_at IS NULL AND expires_at > ?", entityID, time.Now()).
		Order("created_at DESC").
		First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.EntityEmail{}, ErrEmailVerificationNoCode
		}
		return model.EntityEmail{}, err
	}
	if verification.Email != auth.Email {
		return model.EntityEmail{}, ErrEmailVerificationNoCode
	}

	// Count the attempt before comparing, atomically, so parallel guesses
	// cannot exceed the limit.
	result := database.DB.Model(&model.EmailVerification{}).
		Where("id = ? AND attempts < ?", verification.ID, EmailVerificationMaxAttempts).
		Update("attempts", gorm.Expr("attempts + 1"))
	if result.Error != nil {
		return model.EntityEmail{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.EntityEmail{}, ErrEmailVerificationTooManyTry
	}
	expected := hashVerificationCode(verification.ID, code)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(verification.CodeHash)) != 1 {
		if verification.Attempts+1 >= EmailVerificationMaxAttempts {
			return model.EntityEmail{}, ErrEmailVerificationTooManyTry
		}
		return model.EntityEmail{}, ErrEmailVerificationWrongCode
	}

	now := time.Now()
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		consumed := tx.Model(&model.EmailVerification{}).
			Where("id = ? AND consumed_at IS NULL", verification.ID).
			Update("consumed_at", now)
		if consumed.Error != nil {
			return consumed.Error
		}
		if consumed.RowsAffected == 0 {
			return ErrEmailVerificationNoCode
		}
		verified := tx.Model(&model.EntityEmail{}).
			Where("entity_id = ? AND email = ?", entityID, verification.Email).
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

func checkEmailVerificationThrottle(entityID string, email string) error {
	var recent []model.EmailVerification
	if err := database.DB.
		Where("entity_id = ? AND created_at > ?", entityID, time.Now().Add(-time.Hour)).
		Order("created_at DESC").
		Find(&recent).Error; err != nil {
		return err
	}
	if len(recent) == 0 {
		return nil
	}
	for _, verification := range recent {
		if verification.Email != email {
			continue
		}
		if wait := EmailVerificationResendAfter - time.Since(verification.CreatedAt); wait > 0 {
			return EmailVerificationThrottled{RetryAfter: wait}
		}
		break
	}
	if len(recent) >= emailVerificationMaxPerHour {
		oldest := recent[len(recent)-1]
		return EmailVerificationThrottled{RetryAfter: time.Until(oldest.CreatedAt.Add(time.Hour))}
	}
	return nil
}

func sendVerificationEmail(ctx context.Context, entityID string, code string) error {
	minutes := int(EmailVerificationCodeTTL.Minutes())
	var html bytes.Buffer
	if err := emailVerificationTemplate.Execute(&html, map[string]any{"Code": code, "Minutes": minutes}); err != nil {
		return err
	}
	token, err := CoreServiceToken()
	if err != nil {
		return fmt.Errorf("mint core service token: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, emailVerificationNotifyTimeout)
	defer cancel()
	client := notify.Client{BaseURL: config.NotifyURL, Token: token}
	return client.SendEmail(ctx, notify.Email{
		EntityID: entityID,
		SenderID: config.NotifyEmailSenderID,
		Category: "email-verification",
		Subject:  code + " is your Gaucho Racing verification code",
		Body: fmt.Sprintf("Your Gaucho Racing verification code is %s.\n\n"+
			"Enter it in Sentinel to verify the email on your account. It expires in %d minutes.\n\n"+
			"Never share this code with anyone. If you didn't request it, you can ignore this email.", code, minutes),
		HTML: html.String(),
	})
}

func generateVerificationCode() (string, error) {
	limit := big.NewInt(1)
	for range emailVerificationCodeDigits {
		limit.Mul(limit, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", emailVerificationCodeDigits, n), nil
}

func hashVerificationCode(verificationID string, code string) string {
	sum := sha256.Sum256([]byte(verificationID + ":" + code))
	return hex.EncodeToString(sum[:])
}
