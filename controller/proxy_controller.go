package controller

import (
	"fmt"
	"net/http"
	"sentinel/service"
	"sentinel/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func OauthProxyValidate(c *gin.Context) {
	clientID := c.Query("client_id")
	app := service.GetClientApplicationByID(clientID)
	if app.ID == "" {
		utils.SugarLogger.Errorf("Invalid client_id: %s", clientID)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
	}
	originalHost := c.GetHeader("X-Original-Host")
	if originalHost == "" {
		utils.SugarLogger.Infoln("Missing X-Original-Host header")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "you are not authorized to access this resource"})
		return
	}
	hasRedirect := false
	for _, redirect := range app.RedirectURIs {
		if strings.Contains(redirect, originalHost) {
			hasRedirect = true
			break
		}
	}
	if !hasRedirect {
		utils.SugarLogger.Errorf("Invalid X-Original-Host header: %s", originalHost)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}

	// Check last auth time for client application
	lastAuthTime, err := c.Cookie(fmt.Sprintf("sentinel_%s", app.ID))
	if err != nil || lastAuthTime == "" {
		utils.SugarLogger.Infoln("Missing last auth time cookie")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}
	lastAuthTimeInt, err := strconv.Atoi(lastAuthTime)
	if err != nil {
		utils.SugarLogger.Errorf("Invalid last auth time cookie: %s", lastAuthTime)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}
	timeSinceLastAuth := (time.Now().UnixMilli() - int64(lastAuthTimeInt)) / 1000
	utils.SugarLogger.Infof("Last authorized %d seconds ago", timeSinceLastAuth)
	if timeSinceLastAuth > 60*60 {
		utils.SugarLogger.Errorf("Last authorized too long ago")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}

	// Check sentinel access token
	accessToken, err := c.Cookie("sentinel_access_token")
	if err != nil || accessToken == "" {
		utils.SugarLogger.Infoln("Missing sentinel_access_token cookie")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
		return
	}
	claims, err := service.ValidateJWT(accessToken)
	if err != nil {
		utils.SugarLogger.Errorln("Failed to validate token: " + err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
	} else {
		utils.SugarLogger.Infof("Decoded token: %s (%s)", claims.ID, claims.Email)
		utils.SugarLogger.Infof("↳ Client ID: %s", claims.Audience[0])
		utils.SugarLogger.Infof("↳ Scope: %s", claims.Scope)
		utils.SugarLogger.Infof("↳ Issued at: %s", claims.IssuedAt.String())
		utils.SugarLogger.Infof("↳ Expires at: %s", claims.ExpiresAt.String())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Authenticated"})
}
