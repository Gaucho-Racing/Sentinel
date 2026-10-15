package api

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

type confirmEmailVerificationRequest struct {
	Code string `json:"code" binding:"required"`
}

func SendMyEmailVerification(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	sent, err := service.SendEmailVerificationCode(c.Request.Context(), entityID)
	if err != nil {
		writeEmailVerificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, sent)
}

func ConfirmMyEmailVerification(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	var req confirmEmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	auth, err := service.ConfirmEmailVerificationCode(entityID, req.Code)
	if err != nil {
		writeEmailVerificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, auth)
}

func writeEmailVerificationError(c *gin.Context, err error) {
	var throttled service.EmailVerificationThrottled
	switch {
	case errors.As(err, &throttled):
		seconds := int(math.Ceil(throttled.RetryAfter.Seconds()))
		c.Header("Retry-After", strconv.Itoa(seconds))
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "retry_after_seconds": seconds})
	case errors.Is(err, service.ErrEmailAlreadyVerified):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNoEmailOnAccount):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrEmailVerificationWrongCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "wrong_code"})
	case errors.Is(err, service.ErrEmailVerificationNoCode):
		c.JSON(http.StatusGone, gin.H{"error": err.Error(), "code": "no_active_code"})
	case errors.Is(err, service.ErrEmailVerificationTooManyTry):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "too_many_attempts"})
	case errors.Is(err, service.ErrEmailVerificationNotSent):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type reauthenticateRequest struct {
	Password string `json:"password" binding:"required"`
}

func ReauthenticateMe(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	var req reauthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}
	grant, err := service.Reauthenticate(entityID, GetRequestTokenAudience(c), req.Password)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, grant)
	case errors.Is(err, service.ErrIncorrectPassword):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "incorrect_password"})
	case errors.Is(err, service.ErrNoEmailOnAccount):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type changeEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

func ChangeMyEmail(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	var req changeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	reauthTokenID, err := service.CheckReauthToken(entityID, c.GetHeader("X-Reauth-Token"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "reauth_required"})
		return
	}
	auth, err := service.ChangeEmail(entityID, req.Email)
	switch {
	case err == nil:
		service.RevokeReauthToken(reauthTokenID)
		c.JSON(http.StatusOK, auth)
	case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrEmailUnchanged):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrEmailInUse):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNoEmailOnAccount):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type passwordChangeRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
	Code        string `json:"code"`
}

func SendMyPasswordChangeCode(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	var req passwordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
		return
	}
	sent, err := service.SendPasswordChangeCode(c.Request.Context(), entityID, req.NewPassword)
	if err != nil {
		writePasswordChangeError(c, err)
		return
	}
	c.JSON(http.StatusOK, sent)
}

func ChangeMyPassword(c *gin.Context) {
	entityID := GetRequestTokenEntityID(c)
	Require(c, RequestTokenHasFirstPartyAccess(c) && entityID != "")

	var req passwordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_password and code are required"})
		return
	}
	if err := service.ChangePassword(c.Request.Context(), entityID, req.Code, req.NewPassword); err != nil {
		writePasswordChangeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writePasswordChangeError(c *gin.Context, err error) {
	var invalid service.InvalidPasswordError
	switch {
	case errors.As(err, &invalid), errors.Is(err, service.ErrPasswordUnchanged):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_password"})
	case errors.Is(err, service.ErrEmailNotVerified):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		writeEmailVerificationError(c, err)
	}
}

func ListRetiredEmails(c *gin.Context) {
	Require(c, RequestTokenHasInternalAccess(c))
	emails, err := service.ListRetiredEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emails)
}
