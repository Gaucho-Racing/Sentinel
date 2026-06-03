package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
)

// ResolvedSP mirrors core's ResolvedSAMLServiceProvider: the SP registration
// plus the identifying fields of its owning application. ClientID drives the
// access gate and group filtering; the app name/icon feed the consent screen.
type ResolvedSP struct {
	ApplicationID           string `json:"application_id"`
	EntityID                string `json:"entity_id"`
	ACSURL                  string `json:"acs_url"`
	NameIDFormat            string `json:"name_id_format"`
	CertificatePEM          string `json:"certificate_pem"`
	WantAuthnRequestsSigned bool   `json:"want_authn_requests_signed"`
	MetadataXML             string `json:"metadata_xml"`
	ClientID                string `json:"client_id"`
	AppName                 string `json:"app_name"`
	AppIconURL              string `json:"app_icon_url"`
}

// ResolveSP fetches the SP registration for a SAML entityID from core. Returns
// sentinel.APIError (with Status 404) when no SP is registered for the id. The
// entityID is sent in the request body, not the path: SAML entity IDs are
// usually URLs whose `://` and slashes break a path segment.
func ResolveSP(entityID string) (ResolvedSP, error) {
	var sp ResolvedSP
	if err := sentinel.Post("/api/core/saml/sp/resolve", map[string]string{"entity_id": entityID}, &sp); err != nil {
		return ResolvedSP{}, err
	}
	return sp, nil
}

// spProvider implements saml.ServiceProviderProvider, resolving an inbound
// AuthnRequest's issuer to its SP metadata. The request argument is unused —
// resolution is keyed solely on the entityID.
type spProvider struct{}

func (p *spProvider) GetServiceProvider(_ *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	sp, err := ResolveSP(serviceProviderID)
	if err != nil {
		// crewjam requires os.ErrNotExist to distinguish "unknown SP" from a
		// transient lookup failure.
		var apiErr *sentinel.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return sp.entityDescriptor()
}

// entityDescriptor turns a resolved SP into the metadata crewjam needs to
// validate the request and locate the ACS. Published SP metadata (MetadataXML)
// is authoritative when present; otherwise we synthesize a minimal descriptor
// from the discrete fields (entityID + HTTP-POST ACS).
func (sp ResolvedSP) entityDescriptor() (*saml.EntityDescriptor, error) {
	if sp.MetadataXML != "" {
		ed, err := samlsp.ParseMetadata([]byte(sp.MetadataXML))
		if err != nil {
			return nil, fmt.Errorf("parse SP metadata for %s: %w", sp.EntityID, err)
		}
		return ed, nil
	}
	if sp.ACSURL == "" {
		return nil, fmt.Errorf("SP %s has neither metadata nor an ACS URL", sp.EntityID)
	}
	return &saml.EntityDescriptor{
		EntityID: sp.EntityID,
		SPSSODescriptors: []saml.SPSSODescriptor{{
			AssertionConsumerServices: []saml.IndexedEndpoint{{
				Binding:  saml.HTTPPostBinding,
				Location: sp.ACSURL,
				Index:    1,
			}},
		}},
	}, nil
}
