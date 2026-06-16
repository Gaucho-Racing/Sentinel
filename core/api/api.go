package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	api := InitializeRouter()
	InitializeRoutes(api)
	err := api.Run(":" + config.Port)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to start server: %v", err)
	}
}

func InitializeRouter() *gin.Engine {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))
	r.Use(AuthChecker())
	r.Use(UnauthorizedPanicHandler())
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET("/core/ping", Ping)
	router.GET("/core/keys", JWKS)
	router.POST("/core/token", GenerateToken)
	router.POST("/core/token/validate", ValidateToken)
	router.DELETE("/core/token/:id", RevokeToken)

	router.GET("/core/entity/external/:provider", ListExternalAuthsByProvider)
	router.GET("/core/entity/external/:provider/:externalID", GetEntityByExternalAuth)
	router.POST("/core/entity/logins", CreateEntityLogin)
	router.POST("/core/entity", CreateEntity)
	router.GET("/core/entity/:entityID", GetEntityByID)
	router.GET("/core/entity/:entityID/groups", GetEntityGroups)
	router.GET("/core/entity/:entityID/memberships", GetEntityMemberships)
	router.GET("/core/entity/:entityID/logins", GetEntityLogins)
	router.POST("/core/entity/:entityID/email-auth", CreateEntityEmailAuth)
	router.POST("/core/entity/:entityID/phone-auth", CreateEntityPhoneAuth)
	router.POST("/core/entity/:entityID/external-auth", CreateEntityExternalAuth)
	router.PATCH("/core/entity/:entityID/external-auth/:provider", UpdateEntityExternalAuthMetadata)
	router.POST("/core/users", CreateOrUpdateUser)

	router.POST("/core/applications/verify", VerifyClientCredentials)
	router.GET("/core/applications/client/:clientID/groups", GetApplicationGroupsByClientID)
	router.POST("/core/saml/sp/resolve", ResolveSAMLServiceProvider)
	router.POST("/core/login/email-password", LoginEmailPassword)

	router.GET("/entities/@me", GetMe)
	router.GET("/entities/:id", GetEntity)

	router.GET("/users", GetAllUsers)
	router.GET("/users/check-username", CheckUsername)
	router.GET("/users/:id", GetUserByID)
	router.POST("/users", CreateOrUpdateUser)
	router.DELETE("/users/:id", DeleteUser)
	router.GET("/users/:id/groups", GetUserGroups)
	router.GET("/users/:id/logins", GetUserLogins)
	router.GET("/users/:id/recent-applications", GetUserRecentApplications)

	router.GET("/applications", GetAllApplications)
	router.GET("/applications/:id", GetApplicationByID)
	router.GET("/applications/client/:clientID", GetApplicationByClientID)
	router.POST("/applications", CreateApplication)
	router.PUT("/applications/:id", UpdateApplication)
	router.DELETE("/applications/:id", DeleteApplication)
	router.GET("/applications/:id/secret", GetApplicationSecret)
	router.GET("/applications/:id/groups", GetApplicationGroups)
	router.POST("/applications/:id/groups", AddApplicationGroup)
	router.DELETE("/applications/:id/groups/:groupID", RemoveApplicationGroup)
	router.GET("/applications/:id/service-accounts", ListServiceAccountsForApplication)
	router.POST("/applications/:id/service-accounts", CreateServiceAccountForApp)
	router.DELETE("/service-accounts/:id", DeleteServiceAccount)
	router.GET("/service-accounts/:id/api-keys", ListAPIKeysForServiceAccount)
	router.POST("/service-accounts/:id/api-keys", CreateAPIKeyForServiceAccount)
	router.DELETE("/service-accounts/:id/api-keys/:keyID", RevokeAPIKey)

	router.GET("/applications/:id/redirect-uris", GetApplicationRedirectURIs)
	router.POST("/applications/:id/redirect-uris", AddApplicationRedirectURI)
	router.DELETE("/applications/:id/redirect-uris", RemoveApplicationRedirectURI)
	router.GET("/applications/:id/saml", GetApplicationSAML)
	router.POST("/applications/:id/saml", UpsertApplicationSAML)
	router.DELETE("/applications/:id/saml", DeleteApplicationSAML)

	router.GET("/groups", GetAllGroups)
	router.GET("/groups/:id", GetGroupByID)
	router.POST("/groups", CreateOrUpdateGroup)
	router.DELETE("/groups/:id", DeleteGroup)

	router.GET("/groups/:id/applications", GetGroupApplications)

	router.GET("/groups/:id/members", GetGroupMembers)
	router.POST("/groups/:id/members", AddGroupMember)
	router.DELETE("/groups/:id/members/:entityID", RemoveGroupMember)

	router.GET("/groups/:id/conditional-bindings", GetGroupConditionalBindings)
	router.POST("/groups/:id/conditional-bindings", CreateGroupConditionalBinding)
	router.DELETE("/groups/:id/conditional-bindings/:bindingID", DeleteGroupConditionalBinding)

	router.GET("/groups/:id/owners", GetGroupOwners)
	router.POST("/groups/:id/owners", AddGroupOwner)
	router.DELETE("/groups/:id/owners/:entityID", RemoveGroupOwner)

	router.GET("/groups/:id/requests", GetGroupJoinRequests)
	router.GET("/groups/:id/requests/:requestID", GetGroupJoinRequest)
	router.POST("/groups/:id/requests", CreateGroupJoinRequest)
	router.POST("/groups/:id/requests/:requestID/approve", ApproveGroupJoinRequest)
	router.POST("/groups/:id/requests/:requestID/reject", RejectGroupJoinRequest)
	router.DELETE("/groups/:id/requests/:requestID", DeleteGroupJoinRequest)

	router.POST("/groups/:id/requests/:requestID/comments", CreateJoinRequestComment)
	router.DELETE("/groups/:id/requests/:requestID/comments/:commentID", DeleteJoinRequestComment)
}

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.Split(authHeader, "Bearer ")[1]
				// Service-account API keys take the sk_ shape; everything
				// else is treated as a JWT. The two paths set the same
				// Auth-* context values so downstream gates don't have to
				// care which credential type they got.
				if service.HasAPIKeyPrefix(token) {
					authenticateWithAPIKey(c, token)
				} else {
					authenticateWithJWT(c, token)
				}
			}
		}
		c.Next()
	}
}

// authenticateWithJWT validates a bearer token as a Sentinel-issued JWT
// and populates Auth-* context on success. On validation failure aborts
// the request with a 401 — this mirrors the old AuthChecker behavior so
// existing callers see no change.
func authenticateWithJWT(c *gin.Context, token string) {
	claims, err := service.ValidateToken(token)
	if err != nil {
		logger.SugarLogger.Errorln("Failed to validate token: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}
	logger.SugarLogger.Infof("Decoded token: %s (%s)", claims.ID, claims.Subject)
	logger.SugarLogger.Infof("↳ Client ID: %s", claims.Audience[0])
	logger.SugarLogger.Infof("↳ Issued at: %s", claims.IssuedAt.String())
	logger.SugarLogger.Infof("↳ Expires at: %s", claims.ExpiresAt.String())
	c.Set("Auth-Token", token)
	c.Set("Auth-EntityID", claims.Subject)
	c.Set("Auth-Audience", claims.Audience[0])
	c.Set("Auth-Scope", claims.Scope)
	c.Set("Auth-Claims", claims.CustomClaims)
}

// authenticateWithAPIKey validates a sk_ token, resolves the SA + app
// (for the audience claim), and populates Auth-* context to match the
// JWT path. Failure → 401. On success kicks a best-effort last_used_at
// update in a goroutine so the response isn't slowed by an extra write.
func authenticateWithAPIKey(c *gin.Context, token string) {
	key, err := service.ValidateAPIKey(token)
	if err != nil {
		logger.SugarLogger.Errorln("Failed to validate api key: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer"})
		return
	}
	sa, err := service.GetServiceAccountByID(key.ServiceAccountID)
	if err != nil {
		// Key exists but its SA doesn't — orphaned row from a partial
		// delete or a bug. Treat as invalid; the cleanup belongs
		// elsewhere.
		logger.SugarLogger.Errorf("api key %s references missing service account %s: %v", key.ID, key.ServiceAccountID, err)
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer"})
		return
	}
	app, err := service.GetApplicationByID(sa.ApplicationID)
	if err != nil {
		logger.SugarLogger.Errorf("service account %s references missing application %s: %v", sa.ID, sa.ApplicationID, err)
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer"})
		return
	}

	logger.SugarLogger.Infof("Decoded api key: %s (sa=%s entity=%s)", key.ID, sa.ID, sa.EntityID)
	logger.SugarLogger.Infof("↳ Client ID: %s", app.ClientID)
	logger.SugarLogger.Infof("↳ Scope: %s", key.Scope)
	c.Set("Auth-Token", token)
	c.Set("Auth-EntityID", sa.EntityID)
	c.Set("Auth-Audience", app.ClientID)
	c.Set("Auth-Scope", key.Scope)
	c.Set("Auth-Claims", map[string]interface{}{})

	// Background last_used_at update — don't slow the request on the
	// extra write. Fire-and-forget is fine; we recover panics in the
	// service layer.
	go service.MarkAPIKeyUsed(key.ID)
}

func UnauthorizedPanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if err == "Unauthorized" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized to access this resource"})
				} else {
					// Handle other panics
					logger.SugarLogger.Errorf("Unexpected panic: %v", err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.(string)})
				}
			}
		}()
		c.Next()
	}
}

// Require checks if a condition is true, otherwise aborts the request
func Require(c *gin.Context, condition bool) {
	if !condition {
		panic("Unauthorized")
	}
}

// Any checks if any condition is true, otherwise returns false
func Any(conditions ...bool) bool {
	for _, condition := range conditions {
		if condition {
			return true
		}
	}
	return false
}

// All checks if all conditions are true, otherwise returns false
func All(conditions ...bool) bool {
	for _, condition := range conditions {
		if !condition {
			return false
		}
	}
	return true
}

func RequestTokenExists(c *gin.Context) bool {
	_, exists := c.Get("Auth-Token")
	return exists
}

func RequestTokenHasScope(c *gin.Context, scope string) bool {
	scopes := GetRequestTokenScopes(c)
	for _, s := range strings.Split(scopes, " ") {
		if s == scope {
			return true
		}
	}
	return false
}

func RequestTokenHasAudience(c *gin.Context, audience string) bool {
	return GetRequestTokenAudience(c) == audience
}

func RequestTokenHasEntityID(c *gin.Context, entityID string) bool {
	return GetRequestTokenEntityID(c) == entityID
}

func GetRequestToken(c *gin.Context) string {
	token, exists := c.Get("Auth-Token")
	if !exists {
		return ""
	}
	return token.(string)
}

func GetRequestTokenScopes(c *gin.Context) string {
	scopes, exists := c.Get("Auth-Scope")
	if !exists {
		return ""
	}
	return scopes.(string)
}

func GetRequestTokenAudience(c *gin.Context) string {
	audience, exists := c.Get("Auth-Audience")
	if !exists {
		return ""
	}
	return audience.(string)
}

func GetRequestTokenClaims(c *gin.Context) map[string]interface{} {
	claims, exists := c.Get("Auth-Claims")
	if !exists {
		return nil
	}
	return claims.(map[string]interface{})
}

// GetRequestTokenEntityID returns the subject (entity_id) of the bearer that
// AuthChecker resolved, or "" if no valid bearer was presented.
func GetRequestTokenEntityID(c *gin.Context) string {
	id, exists := c.Get("Auth-EntityID")
	if !exists {
		return ""
	}
	return id.(string)
}

// GetRequestTokenUserID returns the user_id custom claim from the bearer, or
// "" if the bearer represents a non-user (service account) or no bearer.
func GetRequestTokenUserID(c *gin.Context) string {
	claims := GetRequestTokenClaims(c)
	if claims == nil {
		return ""
	}
	id, _ := claims["user_id"].(string)
	return id
}

// RequestTokenHasUserID returns true when the bearer's user_id claim equals
// the given userID.
func RequestTokenHasUserID(c *gin.Context, userID string) bool {
	return GetRequestTokenUserID(c) == userID
}

// RequestUserIsAdmin reports whether the bearer's subject entity is a
// member of the Admins group. Used to grant admin-only write access
// without requiring per-resource ownership. Returns false for unauth'd
// requests (entity_id is "").
func RequestUserIsAdmin(c *gin.Context) bool {
	return service.IsAdmin(GetRequestTokenEntityID(c))
}
