package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gin-gonic/gin"
)

func JWKS(c *gin.Context) {
	c.JSON(http.StatusOK, config.RsaPublicKeyJWKS)
}
