package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gaucho-racing/sentinel/github/pkg/sentinel"
)

var errGitHubLinkNotFound = errors.New("GitHub account not linked")

type externalIdentity struct {
	EntityID   string `json:"entity_id"`
	ExternalID string `json:"external_id"`
	Provider   string `json:"provider"`
	Metadata   struct {
		Username string `json:"username"`
	} `json:"metadata"`
}

type sentinelGroup struct {
	Name string `json:"name"`
}

func validatePolicyGroups(ctx context.Context) error {
	var groups []sentinelGroup
	if err := sentinel.Get(ctx, "/api/groups", &groups); err != nil {
		return err
	}
	var members, admins bool
	for _, group := range groups {
		if group.Name == "GithubMembers" {
			members = true
		}
		if group.Name == "GithubAdmins" {
			admins = true
		}
	}
	if !members || !admins {
		return fmt.Errorf("GithubMembers and GithubAdmins groups must exist")
	}
	return nil
}

func verifyUser(ctx context.Context, token string) (string, error) {
	var claims struct {
		Subject    string   `json:"sub"`
		Audience   []string `json:"aud"`
		Scope      string   `json:"scope"`
		UserID     string   `json:"user_id"`
		EntityType string   `json:"entity_type"`
	}
	if err := sentinel.Post(ctx, "/api/core/token/validate", map[string]string{"token": token}, &claims); err != nil {
		return "", err
	}
	if claims.Subject == "" || claims.UserID == "" || claims.EntityType != "USER" {
		return "", fmt.Errorf("user session required")
	}
	validAudience := false
	for _, audience := range claims.Audience {
		if audience == "sentinel" {
			validAudience = true
		}
	}
	firstParty := false
	for _, scope := range strings.Fields(claims.Scope) {
		if scope == "sentinel:all" {
			firstParty = true
		}
	}
	if !validAudience || !firstParty {
		return "", fmt.Errorf("first-party user session required")
	}
	return claims.Subject, nil
}

func identities(ctx context.Context) ([]externalIdentity, error) {
	var identities []externalIdentity
	err := sentinel.Get(ctx, "/api/core/entity/external/GITHUB", &identities)
	return identities, err
}

func groupsForEntity(ctx context.Context, entityID string) ([]sentinelGroup, error) {
	var groups []sentinelGroup
	err := sentinel.Get(ctx, "/api/core/entity/"+url.PathEscape(entityID)+"/groups", &groups)
	return groups, err
}

func linkIdentity(ctx context.Context, entityID string, githubID int64, login string) error {
	return sentinel.Put(ctx, "/api/core/entity/"+url.PathEscape(entityID)+"/github-auth", map[string]string{"external_id": fmt.Sprint(githubID), "login": login}, nil)
}

func unlinkIdentity(ctx context.Context, entityID string, expectedExternalID string) (externalIdentity, error) {
	var identity externalIdentity
	path := "/api/core/entity/" + url.PathEscape(entityID) + "/github-auth/" + url.PathEscape(expectedExternalID)
	err := sentinel.Delete(ctx, path, &identity)
	return identity, err
}

func linkedGitHubIdentity(ctx context.Context, entityID string) (externalIdentity, error) {
	var entity struct {
		ExternalAuths []externalIdentity `json:"external_auths"`
	}
	err := sentinel.Get(ctx, "/api/core/entity/"+url.PathEscape(entityID), &entity)
	if err != nil {
		return externalIdentity{}, err
	}
	for _, auth := range entity.ExternalAuths {
		if auth.EntityID == entityID && auth.Provider == "GITHUB" {
			return auth, nil
		}
	}
	return externalIdentity{}, errGitHubLinkNotFound
}

func identityForGitHubID(ctx context.Context, githubID int64) (bool, error) {
	var entity struct {
		ID string `json:"id"`
	}
	err := sentinel.Get(ctx, "/api/core/entity/external/GITHUB/"+fmt.Sprint(githubID), &entity)
	if err == nil {
		return entity.ID != "", nil
	}
	var apiErr *sentinel.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
		return false, nil
	}
	return false, err
}
