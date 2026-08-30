package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetApplicationConfiguration(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	sp, err := service.GetServiceProvider(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SAML service provider not found"})
			return
		}
		logger.SugarLogger.Errorf("load SAML service provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load SAML configuration"})
		return
	}
	c.JSON(http.StatusOK, sp)
}

func PutApplicationConfiguration(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	var sp model.ServiceProvider
	if err := c.ShouldBindJSON(&sp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sp.ApplicationID = c.Param("id")
	stored, err := service.UpsertServiceProvider(sp)
	if err != nil {
		if errors.Is(err, service.ErrInvalidServiceProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.SugarLogger.Errorf("save SAML service provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save SAML configuration"})
		return
	}
	c.JSON(http.StatusOK, stored)
}

func DeleteApplicationConfiguration(c *gin.Context) {
	if !requireApplicationWrite(c) {
		return
	}
	if err := service.DeleteServiceProvider(c.Param("id")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SAML service provider not found"})
			return
		}
		logger.SugarLogger.Errorf("delete SAML service provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete SAML configuration"})
		return
	}
	c.Status(http.StatusNoContent)
}

type previewRequest struct {
	EntityID      string                `json:"entity_id" binding:"required"`
	Configuration model.ServiceProvider `json:"configuration" binding:"required"`
}

func PreviewApplicationAssertion(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !requireApplicationWrite(c) {
		return
	}
	var req previewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Configuration.ApplicationID = c.Param("id")
	preview, err := service.PreviewAssertionConfiguration(req.Configuration, req.EntityID)
	if err != nil {
		if errors.Is(err, service.ErrNameIDEmpty) || errors.Is(err, service.ErrNameIDEmail) || errors.Is(err, service.ErrInvalidServiceProvider) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		logger.SugarLogger.Errorf("preview SAML assertion: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not resolve assertion preview"})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func requireApplicationWrite(c *gin.Context) bool {
	token := GetRequestToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return false
	}
	route := "/api/applications/" + url.PathEscape(c.Param("id")) + "/write-access"
	err := sentinel.Get(route, nil, map[string]string{"Authorization": "Bearer " + token})
	if err == nil {
		return true
	}
	var apiErr *sentinel.APIError
	if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
		status := apiErr.Status
		if status == http.StatusUnauthorized {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": "you are not authorized to manage this application"})
		return false
	}
	logger.SugarLogger.Errorf("check application write access: %v", err)
	c.JSON(http.StatusBadGateway, gin.H{"error": "could not verify application access"})
	return false
}
