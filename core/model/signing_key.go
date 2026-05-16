package model

import "time"

type SigningKey struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Algorithm     string    `json:"algorithm"`
	PrivateKeyPEM string    `json:"-"`
	PublicKeyPEM  string    `json:"public_key_pem"`
	Active        bool      `json:"active" gorm:"index"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SigningKey) TableName() string {
	return "signing_key"
}
