package service

import (
	"fmt"

	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
)

type emailUnverifiedError struct{}

func (emailUnverifiedError) Error() string {
	return "access denied: verify your email address in Sentinel before signing in to other apps"
}

// Is lets every caller that already denies on ErrAccessDenied deny an
// unverified user the same way.
func (emailUnverifiedError) Is(target error) bool {
	return target == ErrAccessDenied
}

var ErrEmailUnverified error = emailUnverifiedError{}

type emailGateEntity struct {
	Type      string `json:"type"`
	EmailAuth struct {
		VerifiedAt *string `json:"verified_at"`
	} `json:"email_auth"`
}

func checkEmailVerified(entityID string) error {
	var e emailGateEntity
	if err := sentinel.Get("/api/core/entity/"+entityID, &e); err != nil {
		return fmt.Errorf("load entity %s: %w", entityID, err)
	}
	if e.Type == "USER" && e.EmailAuth.VerifiedAt == nil {
		return ErrEmailUnverified
	}
	return nil
}
