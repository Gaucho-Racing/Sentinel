package api

import (
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/oauth/service"
	"github.com/gin-gonic/gin"
)

type tokenRequest struct {
	EntityID  string                 `json:"entity_id"`
	ClientID  string                 `json:"client_id"`
	Scope     string                 `json:"scope"`
	ExpiresIn int                    `json:"expires_in"`
	Claims    map[string]interface{} `json:"claims"`
}

type tokenResponse struct {
	Token   string `json:"token"`
	TokenID string `json:"token_id"`
}

type exchangeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// ExchangeToken handles the OAuth token exchange.
// Supports grant_type=authorization_code and grant_type=refresh_token.
func ExchangeToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeExchange(c)
	case "refresh_token":
		handleRefreshTokenExchange(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
	}
}

func handleAuthorizationCodeExchange(c *gin.Context) {
	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
	}
	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client credentials are required"})
		return
	}

	redirectURI := c.PostForm("redirect_uri")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri is required"})
		return
	}

	// Validate client credentials via core
	var app applicationResponse
	err := sentinel.Get("/applications/client/"+clientID, &app)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}
	// We need the secret to validate — fetch via internal endpoint
	if !validateClientSecret(clientID, clientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}

	authCode, err := service.VerifyAuthorizationCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if authCode.ClientID != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id mismatch"})
		return
	}
	if authCode.RedirectURI != redirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri mismatch"})
		return
	}

	if err := service.CheckAccessGate(authCode.EntityID, clientID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "error_description": err.Error()})
		return
	}

	claims := service.BuildTokenClaims(authCode.EntityID, clientID)

	// Generate access token via core
	accessToken, accessTokenID, err := generateToken(authCode.EntityID, clientID, authCode.Scope, config.AccessTokenTTL, claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	// Generate refresh token via core
	refreshToken, refreshTokenID, err := generateToken(authCode.EntityID, clientID, authCode.Scope+" refresh_token", config.RefreshTokenTTL, claims)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate refresh token: %v", err)
		refreshToken = ""
		refreshTokenID = ""
	}

	sentinel.Post("/core/entity/logins", map[string]string{
		"entity_id":        authCode.EntityID,
		"client_id":        clientID,
		"scope":            authCode.Scope,
		"access_token_id":  accessTokenID,
		"refresh_token_id": refreshTokenID,
		"ip_address":       GetClientIP(c),
	}, nil)

	// OIDC: issue an ID token when the openid scope was granted. auth_time is
	// the moment the user approved consent (when the code was minted).
	var idToken string
	if service.ScopesContain(authCode.Scope, "openid") {
		idClaims := service.BuildIDTokenClaims(authCode.EntityID, clientID, authCode.Scope, authCode.Nonce, accessToken, authCode.CreatedAt.Unix())
		idToken, _, err = generateToken(authCode.EntityID, clientID, authCode.Scope, config.AccessTokenTTL, idClaims)
		if err != nil {
			logger.SugarLogger.Errorf("Failed to generate id token: %v", err)
			idToken = ""
		}
	}

	c.JSON(http.StatusOK, exchangeTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.AccessTokenTTL,
		Scope:        authCode.Scope,
	})
}

func handleRefreshTokenExchange(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
	}
	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client credentials are required"})
		return
	}

	if !validateClientSecret(clientID, clientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}

	// Validate the refresh token via core
	var claims map[string]interface{}
	err := sentinel.Post("/core/token/validate", map[string]string{"token": refreshToken}, &claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	entityID, _ := claims["sub"].(string)
	scope, _ := claims["scope"].(string)

	if !service.ScopesContain(scope, "refresh_token") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "provided token is not a refresh token"})
		return
	}

	// Revoke the old refresh token
	if tokenID, ok := claims["jti"].(string); ok {
		sentinel.Delete("/core/token/"+tokenID, nil)
	}

	// Re-check the gate on refresh — group membership may have changed
	// since the original grant. If the user no longer qualifies, the
	// refresh fails and they have to re-authenticate (which will hit the
	// gate again at the authorize step).
	if err := service.CheckAccessGate(entityID, clientID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "error_description": err.Error()})
		return
	}

	// Strip refresh_token from scope for the access token
	accessScope := service.RemoveScope(scope, "refresh_token")

	newClaims := service.BuildTokenClaims(entityID, clientID)

	// Generate new access token
	accessToken, accessTokenID, err := generateToken(entityID, clientID, accessScope, config.AccessTokenTTL, newClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	// Generate new refresh token (keep refresh_token in scope)
	newRefreshToken, newRefreshTokenID, err := generateToken(entityID, clientID, scope, config.RefreshTokenTTL, newClaims)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to generate refresh token: %v", err)
		newRefreshToken = ""
		newRefreshTokenID = ""
	}

	sentinel.Post("/core/entity/logins", map[string]string{
		"entity_id":        entityID,
		"client_id":        clientID,
		"scope":            accessScope,
		"access_token_id":  accessTokenID,
		"refresh_token_id": newRefreshTokenID,
		"ip_address":       GetClientIP(c),
	}, nil)

	// OIDC: re-issue an ID token on refresh when openid is still in scope. The
	// original nonce isn't replayed on refresh (per spec), and auth_time
	// reflects this refresh since the original authentication time isn't
	// carried forward.
	var idToken string
	if service.ScopesContain(accessScope, "openid") {
		idClaims := service.BuildIDTokenClaims(entityID, clientID, accessScope, "", accessToken, time.Now().Unix())
		idToken, _, err = generateToken(entityID, clientID, accessScope, config.AccessTokenTTL, idClaims)
		if err != nil {
			logger.SugarLogger.Errorf("Failed to generate id token: %v", err)
			idToken = ""
		}
	}

	c.JSON(http.StatusOK, exchangeTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		IDToken:      idToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.AccessTokenTTL,
		Scope:        accessScope,
	})
}

func generateToken(entityID string, clientID string, scope string, expiresIn int, claims map[string]interface{}) (string, string, error) {
	var result tokenResponse
	err := sentinel.Post("/core/token", tokenRequest{
		EntityID:  entityID,
		ClientID:  clientID,
		Scope:     scope,
		ExpiresIn: expiresIn,
		Claims:    claims,
	}, &result)
	if err != nil {
		return "", "", err
	}
	return result.Token, result.TokenID, nil
}


func validateClientSecret(clientID string, clientSecret string) bool {
	var result map[string]interface{}
	err := sentinel.Post("/core/applications/verify", map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	}, &result)
	return err == nil
}
