package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
	"github.com/golang-jwt/jwt/v5"
)

func InitializeKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logger.SugarLogger.Errorln("Failed to generate RSA private key:", err)
		return
	}

	publicKey := &privateKey.PublicKey
	config.RsaPrivateKey = privateKey
	config.RsaPublicKey = publicKey

	config.RsaPublicKeyJWKS = PublicKeyToJWKS(publicKey)

	logger.SugarLogger.Infoln("Successfully generated new RSA keypair")
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
