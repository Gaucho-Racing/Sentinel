package api

import (
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
)

// Metadata serves the IdP's SAML metadata XML (entityID, SSO endpoint, signing
// certificate) for SPs to consume when establishing trust.
func Metadata(c *gin.Context) {
	service.IDP().ServeMetadata(c.Writer, c.Request)
}
