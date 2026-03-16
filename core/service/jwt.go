package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
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

func GenerateAccessToken(entityID string, scope string, clientID string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	entity, _ := GetEntityByID(entityID)
	claims := &model.AccessTokenClaims{
		Entity: entity,
		Scope:  scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ulid.Make().Prefixed("jwt"),
			Subject:   entityID,
			Issuer:    "https://sso.gauchoracing.org",
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(config.RsaPrivateKey)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate access token: %v", err)
		return "", err
	}
	return signedToken, nil
}

func GenerateRefreshToken(entityID string, scope string, clientID string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &model.RefreshTokenClaims{
		Scope: "refresh_token",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ulid.Make().Prefixed("jwt"),
			Subject:   entityID,
			Issuer:    "https://sso.gauchoracing.org",
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(config.RsaPrivateKey)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate refresh token: %v", err)
		return "", err
	}

	refreshToken := &model.RefreshToken{
		ID:        claims.ID,
		EntityID:  entityID,
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: expirationTime,
	}
	if err := database.DB.Create(refreshToken).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to save refresh token: %v", err)
		return "", err
	}
	return signedToken, nil
}

func ValidateAccessToken(token string) (*model.AccessTokenClaims, error) {
	claims := &model.AccessTokenClaims{}

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
	if claims.Scope == "" {
		return nil, fmt.Errorf("token has invalid scope")
	}
	if len(claims.Audience) == 0 {
		return nil, fmt.Errorf("token has invalid audience")
	}
	if claims.Audience[0] != "sentinel" && strings.Contains(claims.Scope, "sentinel:all") {
		return nil, fmt.Errorf("token has unauthorized scope")
	}

	return claims, nil
}

func ValidateRefreshToken(token string) (*model.RefreshTokenClaims, error) {
	claims := &model.RefreshTokenClaims{}

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
	if claims.Scope == "" {
		return nil, fmt.Errorf("token has invalid scope")
	}
	if len(claims.Audience) == 0 {
		return nil, fmt.Errorf("token has invalid audience")
	}

	dbToken := &model.RefreshToken{}
	result := database.DB.Where("id = ?", claims.ID).First(&dbToken)
	if result.Error != nil {
		logger.SugarLogger.Errorf("Failed to find refresh token: %v", result.Error)
		return nil, result.Error
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	return claims, nil
}

func RevokeRefreshToken(id string) error {
	result := database.DB.Where("id = ?", id).Delete(&model.RefreshToken{})
	if result.Error != nil {
		logger.SugarLogger.Errorf("Failed to revoke refresh token: %v", result.Error)
		return result.Error
	}
	return nil
}
