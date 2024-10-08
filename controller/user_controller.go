package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "user:read"),
			RequestUserHasRole(c, "d_admin"),
		),
	))

	result := service.GetAllUsers()
	c.JSON(http.StatusOK, result)
}

func GetUserByID(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "user:read"),
			Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
		),
	))

	result := service.GetUserByID(c.Param("userID"))
	if result.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with given id: " + c.Param("userID")})
		return
	}
	c.JSON(http.StatusOK, result)
}

func GetCurrentUser(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "user:read"),
	))

	user := service.GetUserByID(GetRequestUserID(c))
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with given id: " + GetRequestUserID(c)})
		return
	}
	// insanely stupid override to make singlestore work
	if GetRequestTokenAudience(c) == "quZNfANBcdkW" {
		user.Email = service.GauchoRacingEmailReplace(user.Email)
	}
	c.JSON(http.StatusOK, user)
}

func GetUserInfo(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "user:read"),
	))

	user := service.GetUserByID(GetRequestUserID(c))
	if user.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "No user found with given id: " + GetRequestUserID(c)})
		return
	}
	// insanely stupid override to make singlestore work
	if GetRequestTokenAudience(c) == "quZNfANBcdkW" {
		user.Email = service.GauchoRacingEmailReplace(user.Email)
	}
	claims, _ := service.ValidateJWT(strings.Split(c.GetHeader("Authorization"), "Bearer ")[1])
	userInfo := model.UserInfo{
		Sub:           claims.Subject,
		Name:          user.FirstName + " " + user.LastName,
		GivenName:     user.FirstName,
		FamilyName:    user.LastName,
		Profile:       "https://sso.gauchoracing.com/users/" + user.ID,
		Picture:       user.AvatarURL,
		EmailVerified: true,
		User:          user,
	}
	userInfo.BookstackRoles = append(userInfo.BookstackRoles, "Editor")
	if user.IsInnerCircle() {
		userInfo.BookstackRoles = append(userInfo.BookstackRoles, "Lead")
	}
	if user.IsAdmin() {
		userInfo.BookstackRoles = append(userInfo.BookstackRoles, "Admin")
	}
	c.JSON(http.StatusOK, userInfo)
}

func CreateUser(c *gin.Context) {
	Require(c, All(
		Any(
			RequestTokenHasScope(c, "sentinel:all"),
			RequestTokenHasScope(c, "user:write"),
		),
		Any(
			RequestUserHasID(c, c.Param("userID")),
			RequestUserHasRole(c, "d_admin"),
		),
	))

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
	Require(c, All(
		RequestTokenHasScope(c, "sentinel:all"),
		RequestUserHasRole(c, "d_admin"),
	))

	id := c.Param("id")
	err := service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User with id: " + id + " has been deleted"})
}
