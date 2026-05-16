package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// InitializeKeys loads the active RSA signing key from the signing_key
// table, or generates a fresh keypair and persists it if no active key
// exists. Persistence keeps sessions valid across core restarts.
func InitializeKeys() {
	var stored model.SigningKey
	err := database.DB.Where("active = ?", true).First(&stored).Error
	if err == nil {
		priv, perr := parsePrivateKeyPEM(stored.PrivateKeyPEM)
		if perr != nil {
			logger.SugarLogger.Fatalf("Failed to parse stored signing key %s: %v", stored.ID, perr)
			return
		}
		applyKey(priv)
		logger.SugarLogger.Infof("Loaded signing key %s from db", stored.ID)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.SugarLogger.Fatalf("Failed to load signing key: %v", err)
		return
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to generate RSA private key: %v", err)
		return
	}
	privPEM := encodePrivateKeyPEM(priv)
	pubPEM, err := encodePublicKeyPEM(&priv.PublicKey)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to encode public key: %v", err)
		return
	}
	fresh := model.SigningKey{
		ID:            ulid.Make().Prefixed("sig"),
		Algorithm:     "RS256",
		PrivateKeyPEM: privPEM,
		PublicKeyPEM:  pubPEM,
		Active:        true,
	}
	if err := database.DB.Create(&fresh).Error; err != nil {
		logger.SugarLogger.Fatalf("Failed to persist signing key: %v", err)
		return
	}
	applyKey(priv)
	logger.SugarLogger.Infof("Generated and persisted new signing key %s", fresh.ID)
}

func applyKey(priv *rsa.PrivateKey) {
	config.RsaPrivateKey = priv
	config.RsaPublicKey = &priv.PublicKey
	config.RsaPublicKeyJWKS = PublicKeyToJWKS(&priv.PublicKey)
}

func encodePrivateKeyPEM(priv *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
}

func encodePublicKeyPEM(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})), nil
}

func parsePrivateKeyPEM(s string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func PublicKeyToJWKS(publicKey *rsa.PublicKey) map[string]interface{} {
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())

	return map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": "1",
				"n":   n,
				"e":   e,
			},
		},
	}
}

func GenerateToken(entityID string, clientID string, scope string, expiresIn int, claims map[string]interface{}) (string, string, error) {
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	tokenID := ulid.Make().Prefixed("jwt")
	tokenClaims := &model.TokenClaims{
		Scope:        scope,
		CustomClaims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   entityID,
			Issuer:    "https://sso.gauchoracing.org",
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
	signedToken, err := token.SignedString(config.RsaPrivateKey)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate token: %v", err)
		return "", "", err
	}

	dbToken := &model.Token{
		ID:        tokenID,
		EntityID:  entityID,
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: expirationTime,
	}
	if err := database.DB.Create(dbToken).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to save token: %v", err)
		return "", "", err
	}

	return signedToken, tokenID, nil
}

func ValidateToken(token string) (*model.TokenClaims, error) {
	claims := &model.TokenClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return config.RsaPublicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		logger.SugarLogger.Errorf("Failed to parse token: %v", err)
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	if len(claims.Audience) == 0 {
		return nil, fmt.Errorf("token has invalid audience")
	}

	dbToken := &model.Token{}
	result := database.DB.Where("id = ?", claims.ID).First(&dbToken)
	if result.Error != nil {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

func RevokeToken(id string) error {
	result := database.DB.Where("id = ?", id).Delete(&model.Token{})
	if result.Error != nil {
		logger.SugarLogger.Errorf("Failed to revoke token: %v", result.Error)
		return result.Error
	}
	return nil
}
