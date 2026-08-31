package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBoundedQueryInt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		query         string
		expected      int
		expectedValid bool
	}{
		{name: "default", expected: 30, expectedValid: true},
		{name: "valid", query: "?days=90", expected: 90, expectedValid: true},
		{name: "zero", query: "?days=0"},
		{name: "negative", query: "?days=-1"},
		{name: "over maximum", query: "?days=3661"},
		{name: "not an integer", query: "?days=all"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest("GET", "/analytics"+test.query, nil)

			value, valid := boundedQueryInt(context, "days", 30, 3660)
			if valid != test.expectedValid {
				t.Fatalf("valid = %v, want %v", valid, test.expectedValid)
			}
			if valid && value != test.expected {
				t.Fatalf("value = %d, want %d", value, test.expected)
			}
		})
	}
}
