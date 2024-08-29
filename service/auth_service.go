package service

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"sentinel/config"
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func InitializeKeys() {
	// Parse the RSA public key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(config.RsaPublicKeyString))
	if err != nil {
		utils.SugarLogger.Errorln("Failed to parse RSA public key:", err)
	}
	config.RsaPublicKey = publicKey
	// Parse the RSA private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(config.RsaPrivateKeyString))
	if err != nil {
		utils.SugarLogger.Errorln("Failed to parse RSA private key:", err)
	}
	config.RsaPrivateKey = privateKey
	config.RsaPublicKeyJWKS = PublicKeyToJWKS(publicKey)
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

func RegisterEmailPassword(email string, password string) (string, error) {
	user := GetUserByEmail(email)
	if user.ID == "" {
		return "", fmt.Errorf("user does not exist")
	}
	hash := GetPasswordForEmail(email)
	if hash != "" {
		return "", fmt.Errorf("email/password already registered")
	}
	err := ValidatePassword(password)
	if err != nil {
		return "", err
	}
	hash, err = HashPassword(password)
	if err != nil {
		return "", err
	}
	CreateUserAuth(model.UserAuth{
		ID:       user.ID,
		Email:    email,
		Password: hash,
	})
	token, err := GenerateJWT(user.ID, email, "sentinel:all", "sentinel")
	if err != nil {
		return "", err
	}
	return token, nil
}

func RemovePasswordForEmail(email string) error {
	result := database.DB.Table("user_auth").Where("email = ?", email).Delete(&model.UserAuth{})
	return result.Error
}

func LoginEmailPassword(email string, password string) (string, error) {
	user := GetUserByEmail(email)
	if user.ID == "" {
		return "", fmt.Errorf("user does not exist")
	}
	hash := GetPasswordForEmail(email)
	if hash == "" {
		return "", fmt.Errorf("email/password login does not exist, please login with discord")
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
		return "", err
	}
	token, err := GenerateJWT(user.ID, email, "sentinel:all", "sentinel")
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetUserIDFromDiscordCode(code string) (string, error) {
	accessToken, err := ExchangeCodeForToken(code)
	if err != nil {
		return "", err
	}
	user, err := GetDiscordUserFromToken(accessToken.AccessToken)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
		return "", err
	}
	return string(hash), nil
}

func GenerateJWT(id string, email string, scope string, client_id string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &model.AuthClaims{
		Email: email,
		Scope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Issuer:    "https://sso.gauchoracing.com/",
			Audience:  jwt.ClaimStrings{client_id},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(config.RsaPrivateKey)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(token string) (*model.AuthClaims, error) {
	claims := &model.AuthClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return config.RsaPublicKey, nil
	})
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
		return nil, err
	}
	if !ValidateScope(claims.Scope) {
		return nil, fmt.Errorf("token has invalid scope")
	}
	if len(claims.Audience) == 0 {
		return nil, fmt.Errorf("token has invalid audience")
	}
	if claims.Audience[0] != "sentinel" {
		if GetClientApplicationByID(claims.Audience[0]).ID == "" {
			return nil, fmt.Errorf("token has invalid audience")
		}
	}
	if claims.Audience[0] != "sentinel" && strings.Contains(claims.Scope, "sentinel:all") {
		return nil, fmt.Errorf("token has unauthorized scope")
	}
	return claims, nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 64 {
		return fmt.Errorf("password must be at most 64 characters")
	}
	hasNumber := false
	hasCapital := false
	for _, char := range password {
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if unicode.IsUpper(char) {
			hasCapital = true
		}
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasCapital {
		return fmt.Errorf("password must contain at least one capital letter")
	}
	return nil
}

func GetPasswordForEmail(email string) string {
	var password string
	database.DB.Table("user_auth").Where("email = ?", email).Select("password").Scan(&password)
	return password
}

func GetUserAuthByID(id string) model.UserAuth {
	var userAuth model.UserAuth
	database.DB.Where("id = ?", id).First(&userAuth)
	return userAuth
}

func GetUserAuthByEmail(email string) model.UserAuth {
	var userAuth model.UserAuth
	database.DB.Where("email = ?", email).First(&userAuth)
	return userAuth
}

func CreateUserAuth(userAuth model.UserAuth) {
	if database.DB.Where("id = ?", userAuth.ID).Updates(&userAuth).RowsAffected == 0 {
		database.DB.Create(&userAuth)
	} else {
		utils.SugarLogger.Infof("UserAuth with id: %s has been updated!", userAuth.ID)
	}
}
