package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
	"github.com/gaucho-racing/sentinel/oauth/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

func TestWriteBearerValidationErrorDistinguishesInvalidTokenFromUpstreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Init(true)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "invalid token",
			err:        &sentinel.APIError{Status: http.StatusUnauthorized},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "core failure",
			err:        &sentinel.APIError{Status: http.StatusInternalServerError},
			wantStatus: http.StatusBadGateway,
		},
		{
			name:       "transport failure",
			err:        &sentinel.APIError{Err: errors.New("connection refused")},
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)

			writeBearerValidationError(context, test.err)

			if response.Code != test.wantStatus {
				t.Fatalf("expected %d, got %d: %s", test.wantStatus, response.Code, response.Body.String())
			}
		})
	}
}
