package model

import "time"

type ServiceAccount struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EntityID      string    `json:"entity_id" gorm:"index"`
	ApplicationID string    `json:"application_id" gorm:"index"`
	Name          string    `json:"name"`
	// Scope and TTLDays persist the SA's issuance config so rotation can
	// re-mint a token with the same settings. Scope is space-separated
	// (matches JWT scope claim); TTLDays = 0 means tokens never expire
	// (issued with a ~100-year exp claim).
	Scope     string  `json:"scope"`
	TTLDays   int     `json:"ttl_days"`
	CreatedBy string  `json:"created_by"`
	Groups    []Group `json:"groups" gorm:"-"`
	// ActiveToken is the SA's current bearer JWT row (auth_token) — or nil
	// if the token has been deleted (revoked, never minted, or expired
	// past cleanup). Populated by service.PopulateServiceAccount; never
	// stored. The raw JWT string is only exposed on create/rotate.
	ActiveToken *Token    `json:"active_token,omitempty" gorm:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ServiceAccount) TableName() string {
	return "service_account"
}
