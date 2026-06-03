package model

import "time"

// SigningKey is the IdP's RSA keypair and the self-signed X.509 certificate
// that wraps it. Unlike core (which signs JWTs with a bare RSA key), SAML
// signatures must carry an X.509 cert so relying parties can verify them, so
// the saml service owns its own key + cert independent of core's JWT key.
// Persisted so the IdP's certificate — published in metadata and trusted by
// every SP — survives restarts.
type SigningKey struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Algorithm      string    `json:"algorithm"`
	PrivateKeyPEM  string    `json:"-"`
	CertificatePEM string    `json:"certificate_pem"`
	Active         bool      `json:"active" gorm:"index"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SigningKey) TableName() string {
	return "saml_signing_key"
}
