package api

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
)

func TestWriteMetadataReturnsDownloadableXML(t *testing.T) {
	descriptor := &saml.EntityDescriptor{
		EntityID: "https://sso.gauchoracing.com/saml/metadata",
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						KeyDescriptors: []saml.KeyDescriptor{
							{Use: "signing"},
							{Use: "encryption"},
						},
					},
				},
			},
		},
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeMetadata(context, descriptor)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/samlmetadata+xml" {
		t.Fatalf("content type: got %q", got)
	}
	if got := recorder.Header().Get("Content-Disposition"); got != `attachment; filename="sentinel-idp-metadata.xml"` {
		t.Fatalf("content disposition: got %q", got)
	}
	if !strings.HasPrefix(recorder.Body.String(), xml.Header) {
		t.Fatal("metadata does not start with an XML declaration")
	}

	var parsed saml.EntityDescriptor
	if err := xml.Unmarshal(recorder.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	if parsed.EntityID != descriptor.EntityID {
		t.Fatalf("entity ID: got %q, want %q", parsed.EntityID, descriptor.EntityID)
	}
	if len(parsed.IDPSSODescriptors) != 1 {
		t.Fatalf("IDP descriptors: got %d, want 1", len(parsed.IDPSSODescriptors))
	}
	role := parsed.IDPSSODescriptors[0]
	if len(role.KeyDescriptors) != 1 || role.KeyDescriptors[0].Use != "signing" {
		t.Fatalf("key descriptors: got %#v", role.KeyDescriptors)
	}
	if len(role.NameIDFormats) != 3 {
		t.Fatalf("NameID formats: got %d, want 3", len(role.NameIDFormats))
	}
}
