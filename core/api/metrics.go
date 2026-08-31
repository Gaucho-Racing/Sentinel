package api

import (
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/core/observability"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var prometheusHandler = promhttp.Handler()

func AnalyticsMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/analytics/") {
			c.Next()
			return
		}
		started := time.Now()
		defer func() {
			observability.ObserveAnalyticsRequest(c.FullPath(), c.Writer.Status(), time.Since(started))
		}()
		c.Next()
	}
}

func Metrics(c *gin.Context) {
	Require(c, RequestTokenHasInternalAccess(c))
	prometheusHandler.ServeHTTP(c.Writer, c.Request)
}
