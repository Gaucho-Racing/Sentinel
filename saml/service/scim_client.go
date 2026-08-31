package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	scimCoreUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimCoreGroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
	scimPatchSchema     = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

type SCIMClient struct {
	endpoint    string
	accessToken string
	httpClient  *http.Client
}

type SCIMError struct {
	StatusCode int
	Status     string
	Detail     string
	RequestID  string
}

func (e *SCIMError) Error() string {
	message := fmt.Sprintf("SCIM request returned %s", e.Status)
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	if e.RequestID != "" {
		message += " (request " + e.RequestID + ")"
	}
	return message
}

type SCIMUserName struct {
	Formatted  string `json:"formatted"`
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMUser struct {
	Schemas     []string     `json:"schemas,omitempty"`
	ID          string       `json:"id,omitempty"`
	ExternalID  string       `json:"externalId,omitempty"`
	UserName    string       `json:"userName"`
	Name        SCIMUserName `json:"name"`
	DisplayName string       `json:"displayName"`
	Active      bool         `json:"active"`
	Emails      []SCIMEmail  `json:"emails"`
}

type SCIMMember struct {
	Value string `json:"value"`
}

type SCIMGroup struct {
	Schemas     []string     `json:"schemas,omitempty"`
	ID          string       `json:"id,omitempty"`
	ExternalID  string       `json:"externalId,omitempty"`
	DisplayName string       `json:"displayName"`
	Members     []SCIMMember `json:"members,omitempty"`
}

type scimList[T any] struct {
	TotalResults int    `json:"totalResults"`
	StartIndex   int    `json:"startIndex"`
	ItemsPerPage int    `json:"itemsPerPage"`
	NextCursor   string `json:"nextCursor"`
	Resources    []T    `json:"Resources"`
}

type scimPatchOperation struct {
	Operation string `json:"op"`
	Path      string `json:"path,omitempty"`
	Value     any    `json:"value"`
}

type scimPatchRequest struct {
	Schemas    []string             `json:"schemas"`
	Operations []scimPatchOperation `json:"Operations"`
}

func NewSCIMClient(endpoint, accessToken string) *SCIMClient {
	return &SCIMClient{
		endpoint:    strings.TrimRight(endpoint, "/"),
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *SCIMClient) Test(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/ServiceProviderConfig", nil, nil)
}

func (c *SCIMClient) FindUser(ctx context.Context, attribute, value string) (*SCIMUser, error) {
	users, err := c.listUsers(ctx, attribute+" eq "+strconv.Quote(value))
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("SCIM returned multiple users for %s", attribute)
	}
	return &users[0], nil
}

func (c *SCIMClient) ListUsersForGroup(ctx context.Context, groupID string) ([]SCIMUser, error) {
	filter := "groups.value eq " + strconv.Quote(groupID)
	users := []SCIMUser{}
	cursor := ""
	for {
		query := url.Values{}
		query.Set("count", "100")
		query.Set("filter", filter)
		path := "/Users?cursor&" + query.Encode()
		if cursor != "" {
			query.Set("cursor", cursor)
			path = "/Users?" + query.Encode()
		}
		result := scimList[SCIMUser]{}
		if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
			return nil, err
		}
		users = append(users, result.Resources...)
		if result.NextCursor == "" {
			return users, nil
		}
		cursor = result.NextCursor
	}
}

func (c *SCIMClient) listUsers(ctx context.Context, filter string) ([]SCIMUser, error) {
	result := scimList[SCIMUser]{}
	path := "/Users?filter=" + url.QueryEscape(filter)
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result.Resources, nil
}

func (c *SCIMClient) CreateUser(ctx context.Context, user SCIMUser) (SCIMUser, error) {
	user.Schemas = []string{scimCoreUserSchema}
	var created SCIMUser
	if err := c.do(ctx, http.MethodPost, "/Users", user, &created); err != nil {
		return SCIMUser{}, err
	}
	return created, nil
}

func (c *SCIMClient) UpdateUser(ctx context.Context, userID string, user SCIMUser) error {
	request := scimPatchRequest{
		Schemas: []string{scimPatchSchema},
		Operations: []scimPatchOperation{
			{Operation: "replace", Path: "externalId", Value: user.ExternalID},
			{Operation: "replace", Path: "userName", Value: user.UserName},
			{Operation: "replace", Path: "name.givenName", Value: user.Name.GivenName},
			{Operation: "replace", Path: "name.familyName", Value: user.Name.FamilyName},
			{Operation: "replace", Path: "name.formatted", Value: user.Name.Formatted},
			{Operation: "replace", Path: "displayName", Value: user.DisplayName},
			{Operation: "replace", Path: "emails", Value: user.Emails},
			{Operation: "replace", Path: "active", Value: user.Active},
		},
	}
	return c.do(ctx, http.MethodPatch, "/Users/"+url.PathEscape(userID), request, nil)
}

func (c *SCIMClient) SetUserActive(ctx context.Context, userID string, active bool) error {
	request := scimPatchRequest{
		Schemas:    []string{scimPatchSchema},
		Operations: []scimPatchOperation{{Operation: "replace", Path: "active", Value: active}},
	}
	return c.do(ctx, http.MethodPatch, "/Users/"+url.PathEscape(userID), request, nil)
}

func (c *SCIMClient) FindGroup(ctx context.Context, attribute, value string) (*SCIMGroup, error) {
	result := scimList[SCIMGroup]{}
	path := "/Groups?filter=" + url.QueryEscape(attribute+" eq "+strconv.Quote(value))
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	if len(result.Resources) == 0 {
		return nil, nil
	}
	if len(result.Resources) > 1 {
		return nil, fmt.Errorf("SCIM returned multiple groups for %s", attribute)
	}
	return &result.Resources[0], nil
}

func (c *SCIMClient) CreateGroup(ctx context.Context, group SCIMGroup) (SCIMGroup, error) {
	group.Schemas = []string{scimCoreGroupSchema}
	var created SCIMGroup
	if err := c.do(ctx, http.MethodPost, "/Groups", group, &created); err != nil {
		return SCIMGroup{}, err
	}
	return created, nil
}

func (c *SCIMClient) UpdateGroup(ctx context.Context, groupID, externalID, displayName string) error {
	request := scimPatchRequest{
		Schemas: []string{scimPatchSchema},
		Operations: []scimPatchOperation{
			{Operation: "replace", Path: "externalId", Value: externalID},
			{Operation: "replace", Path: "displayName", Value: displayName},
		},
	}
	return c.do(ctx, http.MethodPatch, "/Groups/"+url.PathEscape(groupID), request, nil)
}

func (c *SCIMClient) AddGroupMembers(ctx context.Context, groupID string, userIDs []string) error {
	return c.patchGroupMembers(ctx, groupID, "add", userIDs)
}

func (c *SCIMClient) RemoveGroupMembers(ctx context.Context, groupID string, userIDs []string) error {
	return c.patchGroupMembers(ctx, groupID, "remove", userIDs)
}

func (c *SCIMClient) patchGroupMembers(ctx context.Context, groupID, operation string, userIDs []string) error {
	for start := 0; start < len(userIDs); start += 100 {
		end := min(start+100, len(userIDs))
		members := make([]SCIMMember, 0, end-start)
		for _, userID := range userIDs[start:end] {
			members = append(members, SCIMMember{Value: userID})
		}
		request := scimPatchRequest{
			Schemas: []string{scimPatchSchema},
			Operations: []scimPatchOperation{{
				Operation: operation,
				Path:      "members",
				Value:     members,
			}},
		}
		if err := c.do(ctx, http.MethodPatch, "/Groups/"+url.PathEscape(groupID), request, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *SCIMClient) do(ctx context.Context, method, path string, body, result any) error {
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode SCIM request: %w", err)
		}
	}

	for attempt := 0; attempt < 3; attempt++ {
		request, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, bytes.NewReader(encoded))
		if err != nil {
			return fmt.Errorf("create SCIM request: %w", err)
		}
		request.Header.Set("Authorization", "Bearer "+c.accessToken)
		request.Header.Set("Accept", "application/scim+json")
		if body != nil {
			request.Header.Set("Content-Type", "application/scim+json")
		}

		response, requestErr := c.httpClient.Do(request)
		if requestErr != nil {
			if method == http.MethodGet && attempt < 2 && !errors.Is(requestErr, context.Canceled) {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
				continue
			}
			return fmt.Errorf("send SCIM request: %w", requestErr)
		}

		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read SCIM response: %w", readErr)
		}
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			if result != nil && len(responseBody) > 0 {
				if err := json.Unmarshal(responseBody, result); err != nil {
					return fmt.Errorf("decode SCIM response: %w", err)
				}
			}
			return nil
		}

		if method == http.MethodGet && attempt < 2 && (response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500) {
			delay := time.Duration(attempt+1) * 500 * time.Millisecond
			if retryAfter, parseErr := strconv.Atoi(response.Header.Get("Retry-After")); parseErr == nil && retryAfter > 0 {
				delay = min(time.Duration(retryAfter)*time.Second, 10*time.Second)
			}
			time.Sleep(delay)
			continue
		}

		var failure struct {
			Detail string `json:"detail"`
		}
		_ = json.Unmarshal(responseBody, &failure)
		return &SCIMError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Detail:     failure.Detail,
			RequestID:  response.Header.Get("x-amzn-RequestId"),
		}
	}
	return errors.New("SCIM request failed after retries")
}
