package service

import (
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestUnlinkedMembersWithRoles(t *testing.T) {
	members := []*discordgo.Member{
		nil,
		{User: nil, Roles: []string{"role"}},
		{User: &discordgo.User{ID: "linked"}, Roles: []string{"role"}},
		{User: &discordgo.User{ID: "bot", Bot: true}, Roles: []string{"role"}},
		{User: &discordgo.User{ID: "unroled"}},
		{User: &discordgo.User{ID: "unlinked"}, Roles: []string{"role"}},
	}
	candidates := unlinkedMembersWithRoles(members, map[string]struct{}{"linked": {}})
	if len(candidates) != 1 || candidates[0].User.ID != "unlinked" {
		t.Fatalf("unexpected role removal candidates: %v", candidates)
	}
}

func TestIsDiscordNotFound(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "missing member", err: &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusNotFound}}, want: true},
		{name: "Discord outage", err: &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusServiceUnavailable}}},
		{name: "missing response", err: &discordgo.RESTError{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDiscordNotFound(tc.err); got != tc.want {
				t.Fatalf("isDiscordNotFound(%v) = %t; want %t", tc.err, got, tc.want)
			}
		})
	}
}
