package model

import "github.com/golang-jwt/jwt/v4"

type AuthClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
