package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetApplicationSAML returns the SAML SP registration for an application.
func GetApplicationSAML(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "applications:read"),
	))
	sp, err := service.GetSAMLServiceProviderByApplicationID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "saml service provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sp)
}

type upsertSAMLRequest struct {
	EntityID                string `json:"entity_id" binding:"required"`
	ACSURL                  string `json:"acs_url"`
	NameIDFormat            string `json:"name_id_format"`
	CertificatePEM          string `json:"certificate_pem"`
	WantAuthnRequestsSigned bool   `json:"want_authn_requests_signed"`
	MetadataXML             string `json:"metadata_xml"`
}

// UpsertApplicationSAML creates or replaces the SAML SP registration attached
// to an application. Same write authorization as the rest of the application's
// configuration (owner, admin, or sentinel:all).
func UpsertApplicationSAML(c *gin.Context) {
	id := c.Param("id")
	app, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, app))

	var req upsertSAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sp, err := service.UpsertSAMLServiceProvider(model.SAMLServiceProvider{
		ApplicationID:           id,
		EntityID:                req.EntityID,
		ACSURL:                  req.ACSURL,
		NameIDFormat:            req.NameIDFormat,
		CertificatePEM:          req.CertificatePEM,
		WantAuthnRequestsSigned: req.WantAuthnRequestsSigned,
		MetadataXML:             req.MetadataXML,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sp)
}

func DeleteApplicationSAML(c *gin.Context) {
	id := c.Param("id")
	app, err := service.GetApplicationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, ApplicationWriteAuthorized(c, app))
	if err := service.DeleteSAMLServiceProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "saml service provider deleted"})
}

type resolveSAMLRequest struct {
	EntityID string `json:"entity_id" binding:"required"`
}

// ResolveSAMLServiceProvider resolves a SAML SP by its entityID. Internal
// (/core) route, unauthenticated, for service-to-service use by the saml
// service when it handles an inbound AuthnRequest — it returns the owning
// application's client_id so the saml service can run the same access gate and
// group filtering OAuth uses.
//
// The entityID is passed in the request body, not the path: SAML entity IDs are
// typically URLs (e.g. https://sp.example.com/metadata) whose `://` and slashes
// break a single-segment path param.
func ResolveSAMLServiceProvider(c *gin.Context) {
	// Internal-only route, used by the SAML service to look up an SP
	// at /saml/sso. SAML service now carries its own SA bearer.
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	var req resolveSAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sp, err := service.GetResolvedSAMLServiceProviderByEntityID(req.EntityID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "saml service provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sp)
}
