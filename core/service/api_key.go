package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"gorm.io/gorm"
)

const (
	apiKeyPrefix       = "sk_"
	apiKeyIDLength     = 32
	apiKeySecretLength = 32
	// base62 alphabet — URL-safe and copy-paste friendly. No look-alikes
	// to remove (0/O/I/l) because keys are machine-generated and
	// machine-pasted; readability across glyphs isn't a real concern.
	base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// ErrInvalidAPIKey is returned by ValidateAPIKey for any failure mode —
// wrong shape, unknown id, bad secret, expired. Collapsing the cases at
// the boundary keeps timing-attack surface uniform and the auth-checker
// response uniform ("invalid bearer").
var ErrInvalidAPIKey = errors.New("invalid api key")

// HasAPIKeyPrefix reports whether `token` is shaped like an API key (vs a
// JWT). Used by the auth middleware to decide which validation path to
// take. Cheap string check, doesn't promise the rest of the format is
// valid — ValidateAPIKey does the real work.
func HasAPIKeyPrefix(token string) bool {
	return strings.HasPrefix(token, apiKeyPrefix)
}

// GenerateAPIKey mints a new key for the service account. Returns the
// persisted row AND the raw sk_..._... string — the raw string is the
// ONLY time the secret half is exposed; subsequent reads only see KeyID.
// expiresAt nil means the key never expires.
func GenerateAPIKey(
	serviceAccountID string,
	name string,
	scope string,
	expiresAt *time.Time,
	createdBy string,
) (model.APIKey, string, error) {
	keyID, err := randomBase62(apiKeyIDLength)
	if err != nil {
		return model.APIKey{}, "", fmt.Errorf("generate key id: %w", err)
	}
	secret, err := randomBase62(apiKeySecretLength)
	if err != nil {
		return model.APIKey{}, "", fmt.Errorf("generate secret: %w", err)
	}

	row := model.APIKey{
		ID:               ulid.Make().Prefixed("ak"),
		ServiceAccountID: serviceAccountID,
		Name:             name,
		KeyID:            keyID,
		HashedSecret:     hashSecret(secret),
		Scope:            scope,
		ExpiresAt:        expiresAt,
		CreatedBy:        createdBy,
	}
	if err := database.DB.Create(&row).Error; err != nil {
		return model.APIKey{}, "", err
	}

	raw := apiKeyPrefix + keyID + "_" + secret
	return row, raw, nil
}

// ValidateAPIKey parses a raw sk_..._... string, looks up the row by KeyID,
// constant-time-compares the secret hash, and checks expiry. Returns the
// row on success; ErrInvalidAPIKey for every failure case.
func ValidateAPIKey(raw string) (model.APIKey, error) {
	if !strings.HasPrefix(raw, apiKeyPrefix) {
		return model.APIKey{}, ErrInvalidAPIKey
	}
	rest := raw[len(apiKeyPrefix):]
	// Two fixed-length parts separated by "_". Reject anything else
	// before we touch the DB.
	if len(rest) != apiKeyIDLength+1+apiKeySecretLength || rest[apiKeyIDLength] != '_' {
		return model.APIKey{}, ErrInvalidAPIKey
	}
	keyID := rest[:apiKeyIDLength]
	secret := rest[apiKeyIDLength+1:]

	var key model.APIKey
	if err := database.DB.Where("key_id = ?", keyID).First(&key).Error; err != nil {
		// Collapse "not found" and "db error" to the same error from the
		// caller's perspective — we don't want a "key doesn't exist" oracle.
		return model.APIKey{}, ErrInvalidAPIKey
	}
	if subtle.ConstantTimeCompare([]byte(key.HashedSecret), []byte(hashSecret(secret))) != 1 {
		return model.APIKey{}, ErrInvalidAPIKey
	}
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return model.APIKey{}, ErrInvalidAPIKey
	}
	return key, nil
}

// MarkAPIKeyUsed updates last_used_at to now. Best-effort; errors are
// logged but not propagated — failing to record a usage event shouldn't
// fail the original request.
func MarkAPIKeyUsed(id string) {
	now := time.Now()
	if err := database.DB.Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error; err != nil {
		logger.SugarLogger.Errorf("api key: failed to mark %s as used: %v", id, err)
	}
}

func ListAPIKeysForServiceAccount(serviceAccountID string) ([]model.APIKey, error) {
	keys := []model.APIKey{}
	if err := database.DB.
		Where("service_account_id = ?", serviceAccountID).
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return []model.APIKey{}, err
	}
	return keys, nil
}

func GetAPIKey(id string) (model.APIKey, error) {
	var key model.APIKey
	if err := database.DB.Where("id = ?", id).First(&key).Error; err != nil {
		return model.APIKey{}, err
	}
	return key, nil
}

// DeleteAPIKey scopes the delete to (id, serviceAccountID) so a tampered
// request can't revoke a key on a different SA.
func DeleteAPIKey(id, serviceAccountID string) error {
	res := database.DB.
		Where("id = ? AND service_account_id = ?", id, serviceAccountID).
		Delete(&model.APIKey{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAllAPIKeysForServiceAccount is called when a SA is deleted so
// outstanding keys can't continue to authenticate. Returns no error if
// there were no keys (a fresh SA being deleted is a normal case).
func DeleteAllAPIKeysForServiceAccount(serviceAccountID string) error {
	return database.DB.
		Where("service_account_id = ?", serviceAccountID).
		Delete(&model.APIKey{}).Error
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// randomBase62 returns a cryptographically-random base62 string of the
// requested length. Uses rejection sampling (rand.Read into a byte buf
// then map to base62) so the distribution is uniform across the alphabet.
func randomBase62(n int) (string, error) {
	out := make([]byte, n)
	// Pull a few extra bytes per call so we don't have to loop frequently
	// — at 62 symbols out of 256 byte values, ~25% of bytes are rejected,
	// so a 2x buffer is comfortably enough for most calls.
	buf := make([]byte, n*2)
	idx := 0
	for idx < n {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) < (256 / 62 * 62) {
				out[idx] = base62Alphabet[int(b)%62]
				idx++
				if idx == n {
					break
				}
			}
		}
	}
	return string(out), nil
}
