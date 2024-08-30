package controller

import (
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func GetAllRolesForUser(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "user:read"),
			Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
		),
	))

	roles := service.GetRolesForUser(c.Param("userID"))
	c.JSON(200, roles)
}

func SetRolesForUser(c *gin.Context) {
	Require(c, All(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestUserHasRole(c, "d_admin"),
	))

	var roles []string
	if err := c.ShouldBindJSON(&roles); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	newRoles := service.SetRolesForUser(c.Param("userID"), roles)
	c.JSON(200, newRoles)
}
