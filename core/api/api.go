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

	router.GET("/users", GetAllUsers)
	router.GET("/users/:id", GetUserByID)
	router.POST("/users", CreateOrUpdateUser)
	router.DELETE("/users/:id", DeleteUser)
	router.GET("/users/:id/groups", GetUserGroups)

	router.GET("/applications", GetAllApplications)
	router.GET("/applications/:id", GetApplicationByID)
	router.POST("/applications", CreateOrUpdateApplication)
	router.DELETE("/applications/:id", DeleteApplication)
	router.GET("/applications/:id/groups", GetApplicationGroups)
	router.POST("/applications/:id/groups", AddApplicationGroup)
	router.DELETE("/applications/:id/groups/:groupID", RemoveApplicationGroup)

	router.GET("/groups", GetAllGroups)
	router.GET("/groups/:id", GetGroupByID)
	router.POST("/groups", CreateOrUpdateGroup)
	router.DELETE("/groups/:id", DeleteGroup)

	router.GET("/groups/:id/members", GetGroupMembers)
	router.POST("/groups/:id/members", AddGroupMember)
	router.DELETE("/groups/:id/members/:entityID", RemoveGroupMember)

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
				claims, err := service.ValidateToken(strings.Split(authHeader, "Bearer ")[1])
				if err != nil {
					logger.SugarLogger.Errorln("Failed to validate token: " + err.Error())
					c.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
				} else {
					logger.SugarLogger.Infof("Decoded token: %s (%s)", claims.ID, claims.Subject)
					logger.SugarLogger.Infof("↳ Client ID: %s", claims.Audience[0])
					logger.SugarLogger.Infof("↳ Issued at: %s", claims.IssuedAt.String())
					logger.SugarLogger.Infof("↳ Expires at: %s", claims.ExpiresAt.String())
					c.Set("Auth-Token", strings.Split(authHeader, "Bearer ")[1])
					c.Set("Auth-EntityID", claims.Subject)
					c.Set("Auth-Audience", claims.Audience[0])
					c.Set("Auth-Scope", claims.Scope)
					c.Set("Auth-Claims", claims.CustomClaims)
				}
			}
		}
		c.Next()
	}
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
