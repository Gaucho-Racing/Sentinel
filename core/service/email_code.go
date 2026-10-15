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

const (
	codePurposeVerifyEmail    = "verify_email"
	codePurposeChangePassword = "change_password"
)

var (
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

type emailContent struct {
	Category  string
	Subject   string
	Heading   string
	Intro     string
	Note      string
	Preheader string
	Body      string
}

//go:embed templates/email.html
var emailTemplates embed.FS

var emailTemplate = template.Must(template.ParseFS(emailTemplates, "templates/email.html"))

func sendEmailCode(ctx context.Context, entityID string, purpose string, email string, content func(code string) emailContent) (EmailVerificationSent, error) {
	if err := checkEmailCodeThrottle(entityID, purpose, email); err != nil {
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
		Purpose:   purpose,
		Email:     email,
		ExpiresAt: now.Add(EmailVerificationCodeTTL),
	}
	verification.CodeHash = hashVerificationCode(verification.ID, code)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.EmailVerification{}).
			Where("entity_id = ? AND purpose = ? AND consumed_at IS NULL", entityID, purpose).
			Update("consumed_at", now).Error; err != nil {
			return err
		}
		return tx.Create(&verification).Error
	})
	if err != nil {
		return EmailVerificationSent{}, err
	}

	if err := sendEmail(ctx, entityID, code, content(code)); err != nil {
		logger.SugarLogger.Errorf("Failed to send %s code to entity %s: %v", purpose, entityID, err)
		database.DB.Delete(&model.EmailVerification{}, "id = ?", verification.ID)
		return EmailVerificationSent{}, ErrEmailVerificationNotSent
	}
	return EmailVerificationSent{
		Email:       email,
		ExpiresAt:   verification.ExpiresAt,
		ResendAfter: verification.CreatedAt.Add(EmailVerificationResendAfter),
	}, nil
}

func checkEmailCode(entityID string, purpose string, email string, code string) (string, error) {
	code = strings.TrimSpace(code)
	var verification model.EmailVerification
	err := database.DB.
		Where("entity_id = ? AND purpose = ? AND consumed_at IS NULL AND expires_at > ?", entityID, purpose, time.Now()).
		Order("created_at DESC").
		First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrEmailVerificationNoCode
		}
		return "", err
	}
	if verification.Email != email {
		return "", ErrEmailVerificationNoCode
	}

	// Count the attempt before comparing, atomically, so parallel guesses
	// cannot exceed the limit.
	result := database.DB.Model(&model.EmailVerification{}).
		Where("id = ? AND attempts < ?", verification.ID, EmailVerificationMaxAttempts).
		Update("attempts", gorm.Expr("attempts + 1"))
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", ErrEmailVerificationTooManyTry
	}
	expected := hashVerificationCode(verification.ID, code)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(verification.CodeHash)) != 1 {
		if verification.Attempts+1 >= EmailVerificationMaxAttempts {
			return "", ErrEmailVerificationTooManyTry
		}
		return "", ErrEmailVerificationWrongCode
	}
	return verification.ID, nil
}

func consumeEmailCode(tx *gorm.DB, verificationID string) error {
	consumed := tx.Model(&model.EmailVerification{}).
		Where("id = ? AND consumed_at IS NULL", verificationID).
		Update("consumed_at", time.Now())
	if consumed.Error != nil {
		return consumed.Error
	}
	if consumed.RowsAffected == 0 {
		return ErrEmailVerificationNoCode
	}
	return nil
}

func checkEmailCodeThrottle(entityID string, purpose string, email string) error {
	var recent []model.EmailVerification
	if err := database.DB.
		Where("entity_id = ? AND purpose = ? AND created_at > ?", entityID, purpose, time.Now().Add(-time.Hour)).
		Order("created_at DESC").
		Find(&recent).Error; err != nil {
		return err
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

func sendEmail(ctx context.Context, entityID string, code string, content emailContent) error {
	var html bytes.Buffer
	if err := emailTemplate.Execute(&html, map[string]any{
		"Heading":   content.Heading,
		"Intro":     content.Intro,
		"Note":      content.Note,
		"Preheader": content.Preheader,
		"Code":      code,
		"Minutes":   int(EmailVerificationCodeTTL.Minutes()),
	}); err != nil {
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
		Category: content.Category,
		Subject:  content.Subject,
		Body:     content.Body,
		HTML:     html.String(),
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
