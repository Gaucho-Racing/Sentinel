package service

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/crewjam/saml"
)

func TestConfiguredAssertionMakerEmitsOnlyConfiguredAttributes(t *testing.T) {
	metadataURL, err := url.Parse("https://idp.example.com/saml/metadata")
	if err != nil {
		t.Fatal(err)
	}
	ssoURL, err := url.Parse("https://idp.example.com/saml/sso")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	spDescriptor := saml.SPSSODescriptor{
		AttributeConsumingServices: []saml.AttributeConsumingService{{
			RequestedAttributes: []saml.RequestedAttribute{{
				Attribute: saml.Attribute{
					Name:       "email",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				},
			}},
		}},
	}
	request := &saml.IdpAuthnRequest{
		IDP: &saml.IdentityProvider{
			Certificate: &x509.Certificate{Raw: []byte{1, 2, 3}},
			MetadataURL: *metadataURL,
			SSOURL:      *ssoURL,
		},
		HTTPRequest:             &http.Request{RemoteAddr: "127.0.0.1"},
		Request:                 saml.AuthnRequest{ID: "request-id", IssueInstant: now},
		ServiceProviderMetadata: &saml.EntityDescriptor{EntityID: "https://sp.example.com/metadata"},
		SPSSODescriptor:         &spDescriptor,
		ACSEndpoint:             &saml.IndexedEndpoint{Location: "https://sp.example.com/acs"},
		Now:                     now,
	}
	configured := saml.Attribute{
		Name:       "configured",
		NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
		Values:     []saml.AttributeValue{{Type: "xs:string", Value: "value"}},
	}
	session := &saml.Session{
		CreateTime:       now,
		NameID:           "driver@example.com",
		NameIDFormat:     string(saml.EmailAddressNameIDFormat),
		UserEmail:        "hidden@example.com",
		CustomAttributes: []saml.Attribute{configured},
	}
	if err := (configuredAssertionMaker{}).MakeAssertion(request, session); err != nil {
		t.Fatalf("MakeAssertion returned an error: %v", err)
	}
	attributes := request.Assertion.AttributeStatements[0].Attributes
	if len(attributes) != 1 || attributes[0].Name != configured.Name {
		t.Fatalf("unexpected attributes: %#v", attributes)
	}
}
