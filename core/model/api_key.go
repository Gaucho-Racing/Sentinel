package model

import "time"

// APIKey is a long-lived opaque credential tied to a ServiceAccount.
//
// The raw key format is "sk_<key_id>_<secret>" — a single string the user
// copies once at creation. KeyID is a 32-char base62 identifier (stored in
// plain text, used as the lookup column) and Secret is a 32-char base62
// secret stored ONLY as a SHA-256 hex hash. On validation we split, find
// the row by KeyID, hash the provided secret, and constant-time-compare.
//
// "Show once": APIs that issue keys return the full sk_ string exactly
// once on POST. Subsequent reads only ever expose the KeyID prefix for
// identification — the secret half is unrecoverable, matching how every
// API-key system the user has touched (Stripe / GitHub / etc) behaves.
type APIKey struct {
	ID               string     `json:"id" gorm:"primaryKey"`
	ServiceAccountID string     `json:"service_account_id" gorm:"index"`
	Name             string     `json:"name"`
	// KeyID is the plain identifier half of the full sk_ token. Unique
	// across all keys so a single lookup resolves the row from a token.
	KeyID        string     `json:"key_id" gorm:"uniqueIndex"`
	HashedSecret string     `json:"-"`
	Scope        string     `json:"scope"`
	// ExpiresAt nil means the key never expires (the user opted out of
	// rotation at create time). Validation treats this as "no expiry."
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	CreatedBy  string     `json:"created_by"`
}

func (APIKey) TableName() string {
	return "api_key"
}
