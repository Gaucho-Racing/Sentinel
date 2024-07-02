package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))

	result := service.GetAllUsers()
	c.JSON(http.StatusOK, result)
}

func GetUserByID(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "read:user"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	result := service.GetUserByID(c.Param("userID"))
	if result.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with given id: " + c.Param("userID")})
		return
	}
	c.JSON(http.StatusOK, result)
}

func CreateUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"), RequestTokenHasScope(c, "write:user"))
	RequireAny(c, RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin"))

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user.ID = c.Param("userID")
	err := service.CreateUser(user, RequestTokenHasScope(c, "sentinel:all"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, service.GetUserByID(user.ID))
}

func DeleteUser(c *gin.Context) {
	RequireAny(c, RequestTokenHasScope(c, "sentinel:all"))
	RequireAny(c, RequestUserHasRole(c, "d_admin"))

	id := c.Param("id")
	err := service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User with id: " + id + " has been deleted"})
}
