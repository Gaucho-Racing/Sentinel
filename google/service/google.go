package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gaucho-racing/sentinel/google/config"
	"github.com/gaucho-racing/sentinel/google/pkg/logger"
	"golang.org/x/oauth2/google"
	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// directorySvc is the Admin SDK Directory client, built once at startup from
// the service-account key with domain-wide delegation. nil when Google sync is
// disabled (no credentials configured).
var directorySvc *directory.Service

// memberEntry is a Google Group member reduced to the fields reconcile needs.
type memberEntry struct {
	Email string
	Role  string // OWNER | MANAGER | MEMBER
}

// InitGoogleClient builds the Directory client from GOOGLE_SERVICE_ACCOUNT,
// impersonating GOOGLE_ADMIN_SUBJECT (domain-wide delegation). A no-op when
// sync is disabled, so the service still boots and serves binding CRUD without
// Google credentials.
func InitGoogleClient() error {
	if !config.GoogleSyncEnabled() {
		logger.SugarLogger.Warnln("google sync disabled: GOOGLE_SERVICE_ACCOUNT / GOOGLE_ADMIN_SUBJECT not set")
		return nil
	}
	jwtConfig, err := google.JWTConfigFromJSON([]byte(config.GoogleServiceAccount), directory.AdminDirectoryGroupMemberScope)
	if err != nil {
		return fmt.Errorf("parse google service account: %w", err)
	}
	jwtConfig.Subject = config.GoogleAdminSubject

	ctx := context.Background()
	svc, err := directory.NewService(ctx, option.WithHTTPClient(jwtConfig.Client(ctx)))
	if err != nil {
		return fmt.Errorf("init directory service: %w", err)
	}
	directorySvc = svc
	logger.SugarLogger.Infof("google sync enabled, impersonating %s", config.GoogleAdminSubject)
	return nil
}

// listGroupMembers returns every member of the Google Group, paginated.
func listGroupMembers(ctx context.Context, groupEmail string) ([]memberEntry, error) {
	var members []memberEntry
	err := directorySvc.Members.List(groupEmail).Context(ctx).Pages(ctx, func(page *directory.Members) error {
		for _, m := range page.Members {
			members = append(members, memberEntry{Email: m.Email, Role: m.Role})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list members of %s: %w", groupEmail, err)
	}
	return members, nil
}

// insertMember adds email to the Google Group as a plain MEMBER. A 409
// (already a member) is treated as success — reconcile is idempotent.
func insertMember(ctx context.Context, groupEmail, email string) error {
	_, err := directorySvc.Members.Insert(groupEmail, &directory.Member{Email: email, Role: "MEMBER"}).Context(ctx).Do()
	if err != nil {
		if isStatus(err, 409) {
			return nil
		}
		return fmt.Errorf("insert %s into %s: %w", email, groupEmail, err)
	}
	return nil
}

// deleteMember removes email from the Google Group. A 404 (not a member) is
// treated as success.
func deleteMember(ctx context.Context, groupEmail, email string) error {
	err := directorySvc.Members.Delete(groupEmail, email).Context(ctx).Do()
	if err != nil {
		if isStatus(err, 404) {
			return nil
		}
		return fmt.Errorf("delete %s from %s: %w", email, groupEmail, err)
	}
	return nil
}

func isStatus(err error, code int) bool {
	var gerr *googleapi.Error
	return errors.As(err, &gerr) && gerr.Code == code
}
