package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

func JWKS(c *gin.Context) {
	c.JSON(http.StatusOK, config.RsaPublicKeyJWKS)
}

type generateTokenRequest struct {
	EntityID  string                 `json:"entity_id" binding:"required"`
	ClientID  string                 `json:"client_id" binding:"required"`
	Scope     string                 `json:"scope" binding:"required"`
	ExpiresIn int                    `json:"expires_in" binding:"required"`
	Claims    map[string]interface{} `json:"claims"`
}

func GenerateToken(c *gin.Context) {
	var req generateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := service.GenerateToken(req.EntityID, req.ClientID, req.Scope, req.ExpiresIn, req.Claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

type validateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func ValidateToken(c *gin.Context) {
	var req validateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := service.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claims)
}

func RevokeToken(c *gin.Context) {
	id := c.Param("id")
	if err := service.RevokeToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token revoked"})
}
