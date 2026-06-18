package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
)

// writeGateError mirrors the oauth service: a genuine denial is 403, any other
// gate-evaluation failure fails closed with 502 rather than letting the login
// through.
func writeGateError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrAccessDenied) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "error_description": err.Error()})
		return
	}
	logger.SugarLogger.Errorf("access gate evaluation failed: %v", err)
	c.JSON(http.StatusBadGateway, gin.H{"error": "server_error", "error_description": "could not verify access"})
}

type validateAuthorizeResponse struct {
	SPEntityID string `json:"sp_entity_id"`
	AppName    string `json:"app_name"`
	AppIconURL string `json:"app_icon_url"`
}

// ValidateAuthorize returns the SP info for the consent screen and, when the
// SPA supplies the active entity, enforces the access gate up front so an
// ineligible user sees a clear error instead of a failed redirect at the SP.
func ValidateAuthorize(c *gin.Context) {
	ssoRequestID := c.Query("sso_request")
	if ssoRequestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sso_request is required"})
		return
	}
	stash, err := service.GetSSORequest(ssoRequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sp, err := service.ResolveSP(stash.SPEntityID)
	if err != nil {
		logger.SugarLogger.Errorf("saml authorize: failed to resolve SP %s: %v", stash.SPEntityID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown service provider"})
		return
	}

	entityID := c.Query("entity_id")
	if entityID != "" {
		// entity_id must match the bearer's subject — otherwise an
		// attacker holding any sso_request stash ID could ask the
		// access gate about any user.
		Require(c, RequestTokenHasEntityID(c, entityID))

		if err := service.CheckAccessGate(entityID, sp.ClientID); err != nil {
			if errors.Is(err, service.ErrAccessDenied) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "app_name": sp.AppName, "app_icon_url": sp.AppIconURL})
				return
			}
			logger.SugarLogger.Errorf("access gate evaluation failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "server_error"})
			return
		}
	}

	c.JSON(http.StatusOK, validateAuthorizeResponse{
		SPEntityID: sp.EntityID,
		AppName:    sp.AppName,
		AppIconURL: sp.AppIconURL,
	})
}

type authorizeRequest struct {
	SSORequest string `json:"sso_request" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
}

// Authorize completes SP-initiated SSO after consent: it consumes the stashed
// request, builds a signed SAML Response for the approved entity, records the
// login for audit, and returns the HTTP-POST binding payload for the SPA to
// auto-submit to the SP's ACS.
func Authorize(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// The bearer's subject becomes the assertion subject — req.EntityID
	// is only honored if it matches. Without this check, possession of
	// any sso_request stash would let a caller mint a signed SAML
	// response for an arbitrary user.
	Require(c, RequestTokenHasEntityID(c, req.EntityID))

	// Peek at the stash without consuming it: we only delete it once the
	// assertion is successfully issued, so a transient failure leaves the
	// handle valid for a retry instead of stranding the user.
	stash, err := service.GetSSORequest(req.SSORequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := service.GenerateResponse([]byte(stash.RequestBuffer), stash.RelayState, req.EntityID, GetClientIP(c), stash.CreatedAt)
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			writeGateError(c, err)
			return
		}
		logger.SugarLogger.Errorf("saml authorize: failed to generate response: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "server_error"})
		return
	}

	service.DeleteSSORequest(req.SSORequest)

	sentinel.Post("/api/core/entity/logins", map[string]string{
		"entity_id":  req.EntityID,
		"client_id":  form.ClientID,
		"scope":      "saml",
		"ip_address": GetClientIP(c),
	}, nil)

	c.JSON(http.StatusOK, gin.H{
		"acs_url":       form.ACSURL,
		"saml_response": form.SAMLResponse,
		"relay_state":   form.RelayState,
	})
}
