package controller

import (
	"net/http"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	if config.Env == "PROD" {
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
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET("/ping", Ping)
	router.POST("/auth/register", RegisterAccount)
	router.POST("/auth/login", LoginAccount)
	router.POST("/auth/login/discord", LoginDiscord)
	router.GET("/oauth/authorize", OauthAuthorize)
	router.POST("/oauth/authorize", OauthAuthorize)
	router.GET("/applications", GetAllClientApplications)
	router.GET("/applications/:appID", GetClientApplicationByID)
	router.POST("/applications", CreateClientApplication)
	router.DELETE("/applications/:appID", DeleteClientApplication)
	router.GET("/users", GetAllUsers)
	router.GET("/users/:userID", GetUserByID)
	router.POST("/users/:userID", CreateUser)
	router.GET("/users/:userID/roles", GetAllRolesForUser)
	router.POST("/users/:userID/roles", SetRolesForUser)
	router.GET("/users/:userID/auth", GetAuthForUser)
	router.GET("/users/:userID/drive", GetDriveStatusForUser)
	router.POST("/users/:userID/drive", AddUserToDrive)
	router.DELETE("/users/:userID/drive", RemoveUserFromDrive)
	router.GET("/users/:userID/github", GetGithubStatusForUser)
	router.POST("/users/:userID/github", AddUserToGithub)
	router.GET("/users/:userID/applications", GetClientApplicationsForUser)
}

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			claims, err := service.ValidateJWT(strings.Split(c.GetHeader("Authorization"), "Bearer ")[1])
			if err != nil {
				utils.SugarLogger.Errorln("Failed to validate token: " + err.Error())
				c.AbortWithStatusJSON(401, gin.H{"message": err.Error()})
			} else {
				utils.SugarLogger.Infoln("Decoded token: " + claims.ID + " " + claims.Email)
				c.Set("Auth-UserID", claims.ID)
				c.Set("Auth-Email", claims.Email)
				c.Set("Auth-Audience", claims.Audience[0])
				c.Set("Auth-Scopes", claims.Scopes)
			}
		}
		c.Next()
	}
}

func RequireAll(c *gin.Context, conditions ...bool) {
	for _, condition := range conditions {
		if !condition {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
			return
		}
	}
}

func RequireAny(c *gin.Context, conditions ...bool) {
	for _, condition := range conditions {
		if condition {
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you are not authorized to access this resource"})
}

func RequestUserHasID(c *gin.Context, id string) bool {
	return GetRequestUserID(c) == id
}

func RequestUserHasEmail(c *gin.Context, email string) bool {
	return GetRequestUserEmail(c) == email
}

func RequestUserHasRole(c *gin.Context, role string) bool {
	user := service.GetUserByID(GetRequestUserID(c))
	return user.HasRole(role)
}

func RequestTokenHasScope(c *gin.Context, scope string) bool {
	scopes := GetRequestTokenScopes(c)
	for _, s := range strings.Split(scopes, "+") {
		if s == scope {
			return true
		}
	}
	return false
}

func RequestTokenHasAudience(c *gin.Context, audience string) bool {
	return GetRequestTokenAudience(c) == audience
}

func GetRequestUserID(c *gin.Context) string {
	id, exists := c.Get("Auth-UserID")
	if !exists {
		return ""
	}
	return id.(string)
}

func GetRequestUserEmail(c *gin.Context) string {
	email, exists := c.Get("Auth-Email")
	if !exists {
		return ""
	}
	return email.(string)
}

func GetRequestTokenScopes(c *gin.Context) string {
	scopes, exists := c.Get("Auth-Scopes")
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
