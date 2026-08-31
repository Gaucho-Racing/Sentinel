package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSCIMClientFindUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/Users" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("filter") != `externalId eq "ent 1"` {
			t.Fatalf("filter = %q", request.URL.Query().Get("filter"))
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		response.Header().Set("Content-Type", "application/scim+json")
		_, _ = response.Write([]byte(`{"totalResults":1,"Resources":[{"id":"aws-user","externalId":"ent 1","userName":"user@example.com"}]}`))
	}))
	defer server.Close()

	user, err := NewSCIMClient(server.URL, "secret").FindUser(context.Background(), "externalId", "ent 1")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil || user.ID != "aws-user" {
		t.Fatalf("user = %#v", user)
	}
}

func TestSCIMClientBatchesGroupMembershipChanges(t *testing.T) {
	requestSizes := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var patch scimPatchRequest
		if err := json.NewDecoder(request.Body).Decode(&patch); err != nil {
			t.Fatal(err)
		}
		if len(patch.Operations) != 1 || patch.Operations[0].Operation != "add" || patch.Operations[0].Path != "members" {
			t.Fatalf("operations = %#v", patch.Operations)
		}
		encoded, err := json.Marshal(patch.Operations[0].Value)
		if err != nil {
			t.Fatal(err)
		}
		var members []SCIMMember
		if err := json.Unmarshal(encoded, &members); err != nil {
			t.Fatal(err)
		}
		requestSizes = append(requestSizes, len(members))
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	userIDs := make([]string, 205)
	for index := range userIDs {
		userIDs[index] = fmt.Sprintf("user-%03d", index)
	}
	if err := NewSCIMClient(server.URL, "secret").AddGroupMembers(context.Background(), "group", userIDs); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(requestSizes) != "[100 100 5]" {
		t.Fatalf("request sizes = %v", requestSizes)
	}
}

func TestSCIMClientPaginatesGroupMembers(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		if request.URL.Query().Get("filter") != `groups.value eq "group"` {
			t.Fatalf("filter = %q", request.URL.Query().Get("filter"))
		}
		if request.URL.Query().Get("count") != "100" {
			t.Fatalf("count = %q", request.URL.Query().Get("count"))
		}
		if requests == 1 {
			if _, present := request.URL.Query()["cursor"]; !present {
				t.Fatal("first request did not enable cursor pagination")
			}
			_, _ = response.Write([]byte(`{"nextCursor":"next","Resources":[{"id":"user-1"}]}`))
			return
		}
		if request.URL.Query().Get("cursor") != "next" {
			t.Fatalf("cursor = %q", request.URL.Query().Get("cursor"))
		}
		_, _ = response.Write([]byte(`{"Resources":[{"id":"user-2"}]}`))
	}))
	defer server.Close()

	users, err := NewSCIMClient(server.URL, "secret").ListUsersForGroup(context.Background(), "group")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || users[0].ID != "user-1" || users[1].ID != "user-2" {
		t.Fatalf("users = %#v", users)
	}
}

func TestSCIMClientReturnsBoundedProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/scim+json")
		response.Header().Set("x-amzn-RequestId", "request-id")
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte(`{"detail":"invalid user"}`))
	}))
	defer server.Close()

	err := NewSCIMClient(server.URL, "super-secret-token").Test(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "invalid user") || !strings.Contains(err.Error(), "request-id") {
		t.Fatalf("error = %q", err.Error())
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Fatal("error exposed the access token")
	}
}
