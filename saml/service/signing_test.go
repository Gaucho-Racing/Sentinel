package service

import (
	"crypto/rsa"
	"crypto/x509"
	"os"
	"sync"
	"testing"

	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/database"
	"github.com/gaucho-racing/sentinel/saml/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSigningKeyNeedsRotation(t *testing.T) {
	tests := []struct {
		generation uint
		want       bool
	}{
		{generation: 0, want: true},
		{generation: 1, want: true},
		{generation: signingKeyGeneration, want: false},
		{generation: signingKeyGeneration + 1, want: false},
	}

	for _, test := range tests {
		key := model.SigningKey{Generation: test.generation}
		if got := signingKeyNeedsRotation(key); got != test.want {
			t.Fatalf("generation %d: got %t, want %t", test.generation, got, test.want)
		}
	}
}

func TestNewSigningKeyUsesCurrentIssuerHost(t *testing.T) {
	originalIssuer := config.Issuer
	config.Issuer = "https://sso.gauchoracing.com"
	defer func() { config.Issuer = originalIssuer }()

	key, err := newSigningKey()
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	if key.Generation != signingKeyGeneration {
		t.Fatalf("generation: got %d, want %d", key.Generation, signingKeyGeneration)
	}
	if !key.Active {
		t.Fatal("new signing key is not active")
	}

	privateKey, err := parsePrivateKeyPEM(key.PrivateKeyPEM)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	certificate, err := parseCertificatePEM(key.CertificatePEM)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	if certificate.Subject.CommonName != "sso.gauchoracing.com" {
		t.Fatalf("common name: got %q", certificate.Subject.CommonName)
	}
	if certificate.SignatureAlgorithm != x509.SHA256WithRSA {
		t.Fatalf("signature algorithm: got %s", certificate.SignatureAlgorithm)
	}
	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("public key type: got %T", certificate.PublicKey)
	}
	if publicKey.N.Cmp(privateKey.N) != 0 || publicKey.E != privateKey.E {
		t.Fatal("certificate public key does not match private key")
	}
}

func TestRotateSigningKey(t *testing.T) {
	dsn := os.Getenv("SAML_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SAML_TEST_DATABASE_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	originalIssuer := config.Issuer
	config.Issuer = "https://sso.gauchoracing.com"
	defer func() { config.Issuer = originalIssuer }()

	legacy, err := newSigningKey()
	if err != nil {
		t.Fatalf("generate legacy signing key: %v", err)
	}
	legacy.ID = "legacy"
	if err := db.Exec(`
		CREATE TABLE saml_signing_key (
			id text PRIMARY KEY,
			algorithm text,
			private_key_pem text,
			certificate_pem text,
			active boolean,
			created_at timestamptz
		)
	`).Error; err != nil {
		t.Fatalf("create legacy signing key table: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO saml_signing_key
			(id, algorithm, private_key_pem, certificate_pem, active, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`, legacy.ID, legacy.Algorithm, legacy.PrivateKeyPEM, legacy.CertificatePEM, true).Error; err != nil {
		t.Fatalf("store legacy signing key: %v", err)
	}
	if err := db.AutoMigrate(&model.SigningKey{}); err != nil {
		t.Fatalf("migrate signing keys: %v", err)
	}
	if err := db.First(&legacy, "id = ?", legacy.ID).Error; err != nil {
		t.Fatalf("load migrated legacy signing key: %v", err)
	}
	if legacy.Generation != 1 {
		t.Fatalf("legacy generation: got %d, want 1", legacy.Generation)
	}

	const replicaCount = 4
	start := make(chan struct{})
	results := make(chan model.SigningKey, replicaCount)
	errors := make(chan error, replicaCount)
	var replicas sync.WaitGroup
	for range replicaCount {
		replicas.Add(1)
		go func() {
			defer replicas.Done()
			<-start
			rotated, err := rotateSigningKey()
			if err != nil {
				errors <- err
				return
			}
			results <- rotated
		}()
	}
	close(start)
	replicas.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Fatalf("rotate signing key concurrently: %v", err)
	}
	rotated := <-results
	if rotated.ID == legacy.ID {
		t.Fatal("rotation reused the legacy key")
	}
	if rotated.Generation != signingKeyGeneration {
		t.Fatalf("generation: got %d, want %d", rotated.Generation, signingKeyGeneration)
	}
	for result := range results {
		if result.ID != rotated.ID {
			t.Fatalf("concurrent rotation created %s and %s", rotated.ID, result.ID)
		}
	}

	var activeKeys []model.SigningKey
	if err := db.Where("active = ?", true).Find(&activeKeys).Error; err != nil {
		t.Fatalf("load active signing keys: %v", err)
	}
	if len(activeKeys) != 1 || activeKeys[0].ID != rotated.ID {
		t.Fatalf("active signing keys: got %#v", activeKeys)
	}

	reused, err := rotateSigningKey()
	if err != nil {
		t.Fatalf("reuse rotated signing key: %v", err)
	}
	if reused.ID != rotated.ID {
		t.Fatalf("second rotation created %s, want %s", reused.ID, rotated.ID)
	}
}
