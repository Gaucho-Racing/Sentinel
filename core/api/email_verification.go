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
