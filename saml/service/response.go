package service

import (
	"net/http"
	"time"

	"github.com/crewjam/saml"
)

// ResponseForm is the HTTP-POST binding payload the SPA auto-submits to the
// SP's ACS: the signed SAML Response (base64), the ACS URL, and the RelayState
// echoed back unchanged.
type ResponseForm struct {
	ACSURL       string `json:"acs_url"`
	SAMLResponse string `json:"saml_response"`
	RelayState   string `json:"relay_state"`
	ClientID     string `json:"-"`
}

// GenerateResponse rebuilds the stashed AuthnRequest, re-runs the access gate,
// builds the session from core, and produces a signed SAML Response for the
// approved entity. The gate is re-checked here (not only at consent) so group
// membership changes between consent and approval can't leak a token.
//
// validatedAt is the time the request was first validated at the SSO endpoint.
// crewjam's Validate re-checks the request's IssueInstant against MaxIssueDelay
// using req.Now, and uses the same req.Now to stamp the assertion's validity
// window. Those want different clocks: the staleness check must use the
// original (already-passed) validation time so a slow consent doesn't trip the
// 90s window, while the assertion must be stamped with the real current time so
// SPs see a fresh window. We anchor Now to validatedAt for Validate, then reset
// it to now before building the response.
//
// remoteAddr is the client IP. crewjam stamps it into the assertion's
// SubjectConfirmationData/SubjectLocality Address, dereferencing
// req.HTTPRequest to read it — so HTTPRequest must be non-nil even though we
// rebuild the request from the stashed buffer rather than a live *http.Request.
func GenerateResponse(requestBuffer []byte, relayState string, entityID string, remoteAddr string, validatedAt time.Time) (ResponseForm, error) {
	req := &saml.IdpAuthnRequest{
		IDP:           idp,
		HTTPRequest:   &http.Request{RemoteAddr: remoteAddr},
		RequestBuffer: requestBuffer,
		RelayState:    relayState,
		Now:           validatedAt,
	}
	if err := req.Validate(); err != nil {
		return ResponseForm{}, err
	}
	// UTC: SAML serializes timestamps as xsd:dateTime and crewjam's layout emits
	// a zone offset for non-UTC times (e.g. -07:00 under the pod's TZ), which
	// strict SPs reject — they require the Zulu "...Z" form.
	req.Now = time.Now().UTC()

	sp, err := ResolveSP(req.Request.Issuer.Value)
	if err != nil {
		return ResponseForm{}, err
	}
	if err := CheckAccessGate(entityID, sp.ClientID); err != nil {
		return ResponseForm{}, err
	}

	session, err := BuildSession(entityID, sp.ClientID)
	if err != nil {
		return ResponseForm{}, err
	}

	maker := idp.AssertionMaker
	if maker == nil {
		maker = saml.DefaultAssertionMaker{}
	}
	if err := maker.MakeAssertion(req, session); err != nil {
		return ResponseForm{}, err
	}

	form, err := req.PostBinding()
	if err != nil {
		return ResponseForm{}, err
	}
	return ResponseForm{
		ACSURL:       form.URL,
		SAMLResponse: form.SAMLResponse,
		RelayState:   form.RelayState,
		ClientID:     sp.ClientID,
	}, nil
}
