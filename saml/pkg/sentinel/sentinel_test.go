package sentinel

import "testing"

func TestHasAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name    string
		headers []map[string]string
		want    bool
	}{
		{name: "none", want: false},
		{name: "other header", headers: []map[string]string{{"X-Request-ID": "123"}}, want: false},
		{name: "canonical", headers: []map[string]string{{"Authorization": "Bearer user"}}, want: true},
		{name: "case insensitive", headers: []map[string]string{{"authorization": "Bearer user"}}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasAuthorizationHeader(test.headers); got != test.want {
				t.Fatalf("hasAuthorizationHeader() = %v, want %v", got, test.want)
			}
		})
	}
}
