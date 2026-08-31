package api

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestTokenHasInternalAccessRequiresServiceAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userContext, _ := gin.CreateTestContext(nil)
	userContext.Set("Auth-Scope", "sentinel:all")
	userContext.Set("Auth-Claims", map[string]interface{}{"user_id": "usr_1"})
	if RequestTokenHasInternalAccess(userContext) {
		t.Fatal("user session must not receive internal access")
	}

	serviceContext, _ := gin.CreateTestContext(nil)
	serviceContext.Set("Auth-Scope", "sentinel:all")
	serviceContext.Set("Auth-Claims", map[string]interface{}{"type": "service_account"})
	if !RequestTokenHasInternalAccess(serviceContext) {
		t.Fatal("internal service account should receive internal access")
	}
}
