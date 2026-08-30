package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	applogger "github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	dsig "github.com/russellhaering/goxmldsig"
	"gorm.io/gorm"
)

const (
	signingKeyGeneration     uint  = 2
	signingKeyRotationLockID int64 = 0x53414d4c5349474e
)

// idp is the configured SAML Identity Provider, built once at startup. It holds
// the signing key + certificate and the providers that resolve service
// providers and user sessions against core.
var idp *saml.IdentityProvider

// IDP returns the process-wide IdentityProvider. Safe after InitializeIDP.
func IDP() *saml.IdentityProvider {
	return idp
}

// InitializeIDP loads the active signing key + cert from the database (or
// generates and persists a fresh self-signed pair on first boot) and assembles
// the IdentityProvider. The cert is published in IdP metadata and is the trust
// anchor for every registered SP, so it must persist across restarts.
func InitializeIDP() {
	priv, cert := loadOrCreateSigningKey()

	metadataURL := mustJoin(config.Issuer, config.MetadataPath)
	ssoURL := mustJoin(config.Issuer, config.SSOPath)

	idp = &saml.IdentityProvider{
		Key:                     priv,
		Certificate:             cert,
		Logger:                  logger.DefaultLogger,
		MetadataURL:             metadataURL,
		SSOURL:                  ssoURL,
		ServiceProviderProvider: &spProvider{},
		AssertionMaker:          configuredAssertionMaker{},
		SignatureMethod:         dsig.RSASHA256SignatureMethod,
	}
}

func loadOrCreateSigningKey() (*rsa.PrivateKey, *x509.Certificate) {
	var stored model.SigningKey
	err := database.DB.Where("active = ?", true).
		Order("generation DESC, created_at DESC").
		First(&stored).Error
	if err == nil && !signingKeyNeedsRotation(stored) {
		return parseStoredSigningKey(stored)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		applogger.SugarLogger.Fatalf("Failed to load saml signing key: %v", err)
	}

	stored, err = rotateSigningKey()
	if err != nil {
		applogger.SugarLogger.Fatalf("Failed to rotate saml signing key: %v", err)
	}
	applogger.SugarLogger.Infof("Activated saml signing key %s generation %d", stored.ID, stored.Generation)
	return parseStoredSigningKey(stored)
}

func rotateSigningKey() (model.SigningKey, error) {
	var selected model.SigningKey
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", signingKeyRotationLockID).Error; err != nil {
			return fmt.Errorf("lock signing key rotation: %w", err)
		}

		err := tx.Where("active = ?", true).
			Order("generation DESC, created_at DESC").
			First(&selected).Error
		if err == nil && !signingKeyNeedsRotation(selected) {
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("lock active signing key: %w", err)
		}

		fresh, err := newSigningKey()
		if err != nil {
			return fmt.Errorf("generate signing key: %w", err)
		}
		if err := tx.Create(&fresh).Error; err != nil {
			return fmt.Errorf("store signing key: %w", err)
		}
		if err := tx.Model(&model.SigningKey{}).
			Where("active = ? AND id <> ?", true, fresh.ID).
			Update("active", false).Error; err != nil {
			return fmt.Errorf("deactivate previous signing key: %w", err)
		}
		selected = fresh
		return nil
	})
	if err != nil {
		return model.SigningKey{}, err
	}
	return selected, nil
}

func signingKeyNeedsRotation(key model.SigningKey) bool {
	return key.Generation < signingKeyGeneration
}

func newSigningKey() (model.SigningKey, error) {
	priv, cert, err := generateSelfSignedKeyPair()
	if err != nil {
		return model.SigningKey{}, err
	}
	return model.SigningKey{
		ID:             ulid.Make().Prefixed("samlsig"),
		Generation:     signingKeyGeneration,
		Algorithm:      "RS256",
		PrivateKeyPEM:  encodePrivateKeyPEM(priv),
		CertificatePEM: encodeCertificatePEM(cert),
		Active:         true,
	}, nil
}

func parseStoredSigningKey(stored model.SigningKey) (*rsa.PrivateKey, *x509.Certificate) {
	priv, privateKeyErr := parsePrivateKeyPEM(stored.PrivateKeyPEM)
	cert, certificateErr := parseCertificatePEM(stored.CertificatePEM)
	if privateKeyErr != nil || certificateErr != nil {
		applogger.SugarLogger.Fatalf(
			"Failed to parse stored saml signing key %s: priv=%v cert=%v",
			stored.ID,
			privateKeyErr,
			certificateErr,
		)
	}
	applogger.SugarLogger.Infof("Loaded saml signing key %s generation %d from db", stored.ID, stored.Generation)
	return priv, cert
}

// generateSelfSignedKeyPair mints a 2048-bit RSA key and a long-lived
// self-signed certificate over it. SAML doesn't validate the cert chain — SPs
// pin the exact certificate from IdP metadata — so a self-signed cert with a
// far-future expiry is the right tool; the security comes from the pin, not a CA.
func generateSelfSignedKeyPair() (*rsa.PrivateKey, *x509.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: issuerHost(), Organization: []string{"Gaucho Racing"}},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, err
	}
	return priv, cert, nil
}

func encodePrivateKeyPEM(priv *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
}

func parsePrivateKeyPEM(s string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, fmt.Errorf("invalid private key PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func encodeCertificatePEM(cert *x509.Certificate) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}))
}

func parseCertificatePEM(s string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, fmt.Errorf("invalid certificate PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func mustJoin(base, path string) url.URL {
	u, err := url.Parse(strings.TrimRight(base, "/") + path)
	if err != nil {
		applogger.SugarLogger.Fatalf("invalid issuer URL %q: %v", base, err)
	}
	return *u
}

func issuerHost() string {
	if u, err := url.Parse(config.Issuer); err == nil && u.Host != "" {
		return u.Host
	}
	return config.Name
}
