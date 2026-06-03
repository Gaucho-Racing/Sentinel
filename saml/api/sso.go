package api

import (
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
)

// SSO is the IdP SSO endpoint (SP-initiated Web SSO). It parses and validates
// the inbound AuthnRequest, stashes it, and redirects the browser to the SPA
// consent page — the SPA holds the first-party session, so authentication and
// consent happen there rather than via a server-rendered login form.
func SSO(c *gin.Context) {
	req, err := saml.NewIdpAuthnRequest(service.IDP(), c.Request)
	if err != nil {
		logger.SugarLogger.Errorf("saml sso: failed to parse request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SAML request"})
		return
	}
	if err := req.Validate(); err != nil {
		logger.SugarLogger.Errorf("saml sso: failed to validate request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SAML request"})
		return
	}

	stash, err := service.GenerateSSORequest(req.Request.Issuer.Value, req.RequestBuffer, req.RelayState)
	if err != nil {
		logger.SugarLogger.Errorf("saml sso: failed to stash request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	redirect := config.Issuer + config.AuthorizePath + "?sso_request=" + url.QueryEscape(stash.ID)
	c.Redirect(http.StatusFound, redirect)
}
