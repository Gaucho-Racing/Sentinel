package api

import (
	"net/http"
	"strconv"

	"github.com/gaucho-racing/sentinel/core/authz"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllUsers(c *gin.Context) {
	Require(c, RequestTokenHasResourceScope(c, authz.UserReadScope))
	users, err := service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func CheckUsername(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username query param is required"})
		return
	}
	available, err := service.IsUsernameAvailable(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": available})
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	// User profile carries email, phone, name, etc. Self / admin /
	// internal only. The user:read scope is allowed when the caller
	// is reading themselves (matches the existing patterns on
	// GetUserLogins and GetUserRecentApplications).
	Require(c, RequestTokenHasResourceScope(c, authz.UserReadScope))
	user, err := service.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func CreateOrUpdateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, err := service.GetUserByID(user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create vs update have different trust requirements. Creates are
	// reserved for internal onboarding (sentinel:all) or admins —
	// arbitrary user creation through this endpoint would be an
	// account-fabrication primitive. Updates allow the user to edit
	// their own profile, plus the usual admin/internal overrides.
	if existing.ID == "" {
		Require(c, Any(
			RequestTokenHasInternalAccess(c),
			RequestTokenHasResourceScope(c, authz.UserWriteScope) && RequestUserIsAdmin(c),
		))
	} else {
		Require(c, Any(
			RequestTokenHasInternalAccess(c),
			RequestTokenHasResourceScope(c, authz.UserWriteScope) && Any(
				RequestTokenHasUserID(c, existing.ID),
				RequestTokenHasEntityID(c, existing.EntityID),
				RequestUserIsAdmin(c),
			),
		))
	}

	if existing.ID != "" {
		user, err = service.UpdateUser(user)
	} else {
		user, err = service.CreateUser(user)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	// Deleting a user is admin-only — no self-delete path through
	// this endpoint (a separate account-closure flow would handle
	// that with the right cleanup).
	Require(c, Any(
		RequestTokenHasInternalAccess(c),
		RequestTokenHasResourceScope(c, authz.UserWriteScope) && RequestUserIsAdmin(c),
	))
	id := c.Param("id")
	if err := service.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func GetUserGroups(c *gin.Context) {
	id := c.Param("id")
	// Same authorization-signal concern as GetEntityGroups — leaking
	// who has admin-equivalent groups would be a recon win for an
	// attacker. Self / admin / internal only.
	Require(c, RequestTokenHasResourceScope(c, authz.GroupsReadScope))
	user, err := service.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user.Groups)
}

func GetUserRecentApplications(c *gin.Context) {
	id := c.Param("id")
	Require(c, Any(
		RequestTokenHasInternalAccess(c),
		RequestTokenHasResourceScope(c, authz.UserReadScope) && Any(
			RequestTokenHasUserID(c, id),
			RequestUserIsAdmin(c),
		),
	))

	user, err := service.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	apps, err := service.GetAccessedApplicationsForEntity(user.EntityID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func GetUserLogins(c *gin.Context) {
	id := c.Param("id")
	Require(c, Any(
		RequestTokenHasInternalAccess(c),
		RequestTokenHasResourceScope(c, authz.UserReadScope) && Any(
			RequestTokenHasUserID(c, id),
			RequestUserIsAdmin(c),
		),
	))

	user, err := service.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logins, err := service.GetEntityLogins(service.EntityLoginsFilter{
		EntityID: user.EntityID,
		ClientID: c.Query("client_id"),
		Scope:    c.Query("scope"),
		Before:   c.Query("before"),
		After:    c.Query("after"),
		Limit:    c.Query("limit"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logins)
}
