package model

import "time"

// SSORequest stashes an inbound, validated AuthnRequest across the SPA consent
// round-trip. The SSO endpoint can't read the first-party session (it's a JWT
// held by the SPA, not a cookie), so it parses + validates the request, stores
// it here, and redirects the browser to the SPA consent page keyed by ID. When
// the user approves, the authorize endpoint reloads it to build and sign the
// assertion. Short-lived and single-use, like an OAuth authorization code.
type SSORequest struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SPEntityID    string    `json:"sp_entity_id"`
	RequestBuffer string    `json:"request_buffer"` // raw AuthnRequest XML
	RelayState    string    `json:"relay_state"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SSORequest) TableName() string {
	return "saml_sso_request"
}
