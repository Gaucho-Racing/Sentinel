package service

import "github.com/crewjam/saml"

type configuredAssertionMaker struct{}

func (configuredAssertionMaker) MakeAssertion(req *saml.IdpAuthnRequest, session *saml.Session) error {
	if err := (saml.DefaultAssertionMaker{}).MakeAssertion(req, session); err != nil {
		return err
	}
	attributes := append([]saml.Attribute(nil), session.CustomAttributes...)
	req.Assertion.AttributeStatements = []saml.AttributeStatement{{Attributes: attributes}}
	return nil
}
