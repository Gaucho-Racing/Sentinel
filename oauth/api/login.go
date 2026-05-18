package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

type sessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	EntityID     string `json:"entity_id"`
}

type emailPasswordLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginEmailPassword(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var req emailPasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var verify struct {
		EntityID string `json:"entity_id"`
	}
	if err := sentinel.Post("/core/login/email-password", req, &verify); err != nil {
		logger.SugarLogger.Errorf("login: upstream failure: %v", err)
		var apiErr *sentinel.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	resp, err := mintFirstPartySession(c, verify.EntityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type refreshSessionRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshSession(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var req refreshSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var claims map[string]interface{}
	if err := sentinel.Post("/core/token/validate", map[string]string{"token": req.RefreshToken}, &claims); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	entityID, _ := claims["sub"].(string)
	scope, _ := claims["scope"].(string)
	if entityID == "" || !service.ScopesContain(scope, "refresh_token") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not a refresh token"})
		return
	}

	if tokenID, ok := claims["jti"].(string); ok {
		sentinel.Delete("/core/token/"+tokenID, nil)
	}

	resp, err := mintFirstPartySession(c, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// First-party tokens are scoped to the Sentinel app itself.
const firstPartyClientID = "sentinel"
const firstPartyAccessScope = "user:read user:write groups:read applications:read"
const firstPartyRefreshScope = firstPartyAccessScope + " refresh_token"

// mintFirstPartySession builds claims, mints access + refresh JWTs, and
// records an entity login for audit. Used by /auth/login and /auth/refresh.
func mintFirstPartySession(c *gin.Context, entityID string) (sessionResponse, error) {
	claims := service.BuildTokenClaims(entityID, firstPartyClientID)

	accessToken, accessTokenID, err := generateToken(entityID, firstPartyClientID, firstPartyAccessScope, config.AccessTokenTTL, claims)
	if err != nil {
		return sessionResponse{}, errors.New("failed to generate access token")
	}

	refreshToken, refreshTokenID, err := generateToken(entityID, firstPartyClientID, firstPartyRefreshScope, config.RefreshTokenTTL, claims)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate refresh token for %s: %v", entityID, err)
		refreshToken = ""
		refreshTokenID = ""
	}

	sentinel.Post("/core/entity/logins", map[string]string{
		"entity_id":        entityID,
		"client_id":        firstPartyClientID,
		"scope":            firstPartyAccessScope,
		"access_token_id":  accessTokenID,
		"refresh_token_id": refreshTokenID,
		"ip_address":       GetClientIP(c),
	}, nil)

	return sessionResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AccessTokenTTL,
		EntityID:     entityID,
	}, nil
}
