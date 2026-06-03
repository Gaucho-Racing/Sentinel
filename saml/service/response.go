package service

import (
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
func GenerateResponse(requestBuffer []byte, relayState string, entityID string) (ResponseForm, error) {
	req := &saml.IdpAuthnRequest{
		IDP:           idp,
		RequestBuffer: requestBuffer,
		RelayState:    relayState,
		Now:           time.Now(),
	}
	if err := req.Validate(); err != nil {
		return ResponseForm{}, err
	}

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
