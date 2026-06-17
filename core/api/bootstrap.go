package api

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/jobs"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// bootstrapHeaderName is the request header non-core services send to
// authenticate to BootstrapToken. A header is cleaner than a body field
// (the secret stays out of access logs that record request bodies, and
// proxies can scrub it generically).
const bootstrapHeaderName = "X-Bootstrap-Secret"

type bootstrapTokenRequest struct {
	// Name must match one of jobs.InternalServiceAccountNames — any
	// other value is rejected, so a leaked bootstrap secret can't be
	// used to harvest tokens for admin-created SAs.
	Name string `json:"name" binding:"required"`
}

type bootstrapTokenResponse struct {
	Token string `json:"token"`
}

// BootstrapToken exchanges the shared INTERNAL_BOOTSTRAP_SECRET for a
// pre-seeded service-account JWT. Used by non-core services at startup
// to acquire their long-lived bearer token; subsequent service-to-service
// traffic carries that bearer in Authorization, so the receiving side
// goes through the normal JWT validation path with zero special-casing.
//
// Fails closed if the secret isn't configured on the server — better to
// 503 than accept any request when the gate hasn't been set up.
func BootstrapToken(c *gin.Context) {
	if config.InternalBootstrapSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "bootstrap secret not configured"})
		return
	}

	provided := c.GetHeader(bootstrapHeaderName)
	// Constant-time compare to avoid leaking the secret length via
	// timing differences. subtle.ConstantTimeCompare needs equal-length
	// inputs, hence the length pre-check.
	if len(provided) != len(config.InternalBootstrapSecret) ||
		subtle.ConstantTimeCompare([]byte(provided), []byte(config.InternalBootstrapSecret)) != 1 {
		logger.SugarLogger.Warnf("bootstrap: rejected request with bad secret from %s", c.ClientIP())
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bootstrap secret"})
		return
	}

	var req bootstrapTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !jobs.IsInternalServiceAccountName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is not on the internal allowlist"})
		return
	}

	sa, err := service.GetServiceAccountByName(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Allowlist hit but the seed never ran — surface as 503 so
			// the caller knows the server is misconfigured rather than
			// thinking the secret was wrong.
			logger.SugarLogger.Errorf("bootstrap: internal SA %s missing from DB; seed didn't run?", req.Name)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "internal service account not seeded"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sa.SignedToken == "" {
		// Best-effort mint to recover from a partial-seed state. If
		// this fails too, the caller will keep retrying and we'll
		// surface the underlying error.
		if _, _, err := service.MintServiceAccountToken(sa); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no active token, mint failed: " + err.Error()})
			return
		}
		// Reload to pick up the freshly-written SignedToken.
		sa, err = service.GetServiceAccountByName(req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	logger.SugarLogger.Infof("bootstrap: issued token to internal service %s (sa=%s)", req.Name, sa.ID)
	c.JSON(http.StatusOK, bootstrapTokenResponse{Token: sa.SignedToken})
}
