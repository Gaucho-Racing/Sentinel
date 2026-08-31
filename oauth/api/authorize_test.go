package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthorizeRejectsCallerSuppliedEntityWithoutBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := InitializeRouter()
	InitializeRoutes(router)

	request := httptest.NewRequest(
		http.MethodPost,
		"/oauth/authorize?client_id=client&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=openid",
		strings.NewReader(`{"entity_id":"ent_victim"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestValidateAuthorizeRequiresBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := InitializeRouter()
	InitializeRoutes(router)

	request := httptest.NewRequest(
		http.MethodGet,
		"/oauth/authorize?client_id=client&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=openid&entity_id=ent_victim",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}
