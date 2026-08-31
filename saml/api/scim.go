package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type putSCIMConfigurationRequest struct {
	Endpoint       string     `json:"endpoint" binding:"required"`
	AccessToken    string     `json:"access_token"`
	TokenExpiresAt *time.Time `json:"token_expires_at"`
	Enabled        bool       `json:"enabled"`
}

func GetSCIMConfiguration(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !requireApplicationWrite(c) {
		return
	}
	configuration, err := service.GetSCIMConfiguration(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SCIM configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load SCIM configuration"})
		return
	}
	c.JSON(http.StatusOK, service.SCIMConfigurationResponse(configuration))
}

func PutSCIMConfiguration(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !requireApplicationWrite(c) {
		return
	}
	var request putSCIMConfigurationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	configuration, err := service.UpsertSCIMConfiguration(c.Param("id"), request.Endpoint, request.AccessToken, request.TokenExpiresAt, request.Enabled)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSCIMConfiguration) || errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.SugarLogger.Errorf("save SCIM configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save SCIM configuration"})
		return
	}
	c.JSON(http.StatusOK, service.SCIMConfigurationResponse(configuration))
}

func DeleteSCIMConfiguration(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	if err := service.DeleteSCIMConfiguration(c.Param("id")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SCIM configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete SCIM configuration"})
		return
	}
	c.Status(http.StatusNoContent)
}

func TestSCIMConfiguration(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	if err := service.TestSCIMConfiguration(c.Request.Context(), c.Param("id")); err != nil {
		logger.SugarLogger.Errorf("test SCIM configuration: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SCIM connection succeeded"})
}

func PreviewSCIMSync(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	result, err := service.PreviewSCIMSync(c.Param("id"))
	if err != nil {
		logger.SugarLogger.Errorf("preview SCIM synchronization: %v", err)
		status := http.StatusBadGateway
		message := "could not load the provisioning scope"
		if errors.Is(err, service.ErrInvalidSCIMConfiguration) {
			status = http.StatusUnprocessableEntity
			message = err.Error()
		}
		c.JSON(status, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, result)
}

func SynchronizeSCIM(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	result, err := service.SynchronizeSCIM(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, service.ErrInvalidSCIMConfiguration) || len(result.ValidationErrors) > 0 {
			status = http.StatusUnprocessableEntity
		}
		if errors.Is(err, service.ErrSCIMSyncInProgress) {
			status = http.StatusConflict
		}
		logger.SugarLogger.Errorf("synchronize SCIM: %v", err)
		c.JSON(status, gin.H{"error": err.Error(), "result": result})
		return
	}
	c.JSON(http.StatusOK, result)
}
