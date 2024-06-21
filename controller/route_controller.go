package controller

import (
	"sentinel/config"
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
	// router.POST("/auth/register", RegisterAccount)
	// router.POST("/auth/login", LoginAccount)
	router.GET("/users", GetAllUsers)
	router.GET("/users/:userID", GetUserByID)
	router.POST("/users/:userID", CreateUser)
}

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// if c.GetHeader("Authorization") != "" {
		// 	claims, err := service.ValidateJWT(strings.Split(c.GetHeader("Authorization"), "Bearer ")[1])
		// 	if err != nil {
		// 		utils.SugarLogger.Errorln("Failed to validate token: " + err.Error())
		// 		if errors.Is(err, jwt.ErrTokenExpired) {
		// 			c.AbortWithStatusJSON(401, gin.H{"message": "Token expired"})
		// 		}
		// 	} else {
		// 		utils.SugarLogger.Infoln("Decoded token: " + claims.ID + " " + claims.Email)
		// 		c.Set("Request-UserID", claims.ID)
		// 	}
		// }
		c.Next()
	}
}
