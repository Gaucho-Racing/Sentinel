package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

type loginEmailPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginEmailPassword verifies an email + password against the stored auth.
// Internal: called by the oauth service from /auth/login/email-password.
// Does not mint tokens — that's oauth's job.
func LoginEmailPassword(c *gin.Context) {
	var req loginEmailPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := service.LoginEmailPassword(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity_id": entity.ID})
}
