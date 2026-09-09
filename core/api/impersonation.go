package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

type impersonateTokenRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	ApplicationID string `json:"application_id" binding:"required"`
}

type impersonateTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenID     string `json:"token_id"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func ImpersonateToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !RequestTokenExists(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !RequestTokenHasScope(c, service.ImpersonationScope) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token lacks sentinel:impersonate"})
		return
	}

	var req impersonateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.ImpersonateUser(GetRequestTokenEntityID(c), req.UserID, req.ApplicationID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrImpersonationCallerNotServiceAccount):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrImpersonationAccessDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrImpersonationUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, service.ErrImpersonationApplicationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	recordAudit(c, model.AuditActionImpersonationTokenIssued, "user", result.UserID, model.JSONMap{
		"application_id": result.ApplicationID,
		"client_id":      result.ClientID,
		"token_id":       result.TokenID,
		"scope":          result.Scope,
		"expires_in":     result.ExpiresIn,
	})
	c.JSON(http.StatusOK, impersonateTokenResponse{
		AccessToken: result.AccessToken,
		TokenID:     result.TokenID,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
		Scope:       result.Scope,
	})
}
