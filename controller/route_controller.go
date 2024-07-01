package controller

import (
	"errors"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
				if errors.Is(err, jwt.ErrTokenExpired) {
					c.AbortWithStatusJSON(401, gin.H{"message": "Token expired"})
				}
			} else {
				utils.SugarLogger.Infoln("Decoded token: " + claims.ID + " " + claims.Email)
				c.Set("Request-UserID", claims.ID)
			}
		}
		c.Next()
	}
}
