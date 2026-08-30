package api

import (
	"encoding/xml"
	"net/http"

	"github.com/crewjam/saml"
	"github.com/gaucho-racing/sentinel/saml/pkg/logger"
	"github.com/gaucho-racing/sentinel/saml/service"
	"github.com/gin-gonic/gin"
)

// Metadata serves the IdP's SAML metadata XML (entityID, SSO endpoint, signing
// certificate) for SPs to consume when establishing trust.
func Metadata(c *gin.Context) {
	descriptor := service.IDP().Metadata()
	role := &descriptor.IDPSSODescriptors[0]
	role.NameIDFormats = []saml.NameIDFormat{
		saml.EmailAddressNameIDFormat,
		saml.PersistentNameIDFormat,
		saml.UnspecifiedNameIDFormat,
	}
	signingKeys := role.KeyDescriptors[:0]
	for _, key := range role.KeyDescriptors {
		if key.Use == "signing" {
			signingKeys = append(signingKeys, key)
		}
	}
	role.KeyDescriptors = signingKeys
	payload, err := xml.MarshalIndent(descriptor, "", "  ")
	if err != nil {
		logger.SugarLogger.Errorf("marshal SAML metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate SAML metadata"})
		return
	}
	c.Data(http.StatusOK, "application/samlmetadata+xml", payload)
}
