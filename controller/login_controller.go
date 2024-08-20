package controller

import (
	"net/http"
	"sentinel/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllLogins(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	logins := service.GetAllLogins()
	c.JSON(http.StatusOK, logins)
}

func GetLoginsForUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "logins:read"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	userID := c.Param("userID")
	if c.Query("count") != "" {
		n, err := strconv.Atoi(c.Query("count"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "count must be a number"})
			return
		}
		logins := service.GetLastNLoginsForUser(userID, n)
		c.JSON(http.StatusOK, logins)
		return
	}
	logins := service.GetLoginsForUser(userID)
	c.JSON(http.StatusOK, logins)
}

func GetLoginsForDestination(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	destination := c.Param("appID")
	if c.Query("count") != "" {
		n, err := strconv.Atoi(c.Query("count"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "count must be a number"})
			return
		}
		logins := service.GetLastNLoginsForDestination(destination, n)
		c.JSON(http.StatusOK, logins)
		return
	}
	logins := service.GetLoginsForDestination(destination)
	c.JSON(http.StatusOK, logins)
}

func GetLoginByID(c *gin.Context) {
	loginID := c.Param("loginID")
	login := service.GetLoginByID(loginID)
	if login.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "no login found with id: " + loginID})
		return
	}

	if !RequestTokenHasScope(c, "sentinel:all") {
		RequireAny(c, RequestTokenHasScope(c, "logins:read"))
		RequireAny(c, RequestUserHasRole(c, "d_admin"), RequestUserHasID(c, login.UserID))
	}

	c.JSON(http.StatusOK, login)
}
