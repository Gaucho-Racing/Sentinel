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
	EntityID string `json:"entity_id" binding:"required"`
	Scope    string `json:"scope" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

func GenerateAccessToken(c *gin.Context) {
	var req generateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := service.GenerateAccessToken(req.EntityID, req.Scope, req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

func GenerateRefreshToken(c *gin.Context) {
	var req generateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := service.GenerateRefreshToken(req.EntityID, req.Scope, req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refresh_token": token})
}

type validateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func ValidateAccessToken(c *gin.Context) {
	var req validateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := service.ValidateAccessToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claims)
}

func ValidateRefreshToken(c *gin.Context) {
	var req validateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := service.ValidateRefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claims)
}

func RevokeRefreshToken(c *gin.Context) {
	id := c.Param("id")
	if err := service.RevokeRefreshToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "refresh token revoked"})
}
