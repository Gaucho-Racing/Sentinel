package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func TestAnalyticsMetricsRecordsRecoveredAuthorizationStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := InitializeRouter()
	router.GET("/analytics/metrics-authorization-test", func(c *gin.Context) {
		Require(c, false)
	})

	before := analyticsRequestCount(t, "/analytics/metrics-authorization-test", "401")
	request := httptest.NewRequest(http.MethodGet, "/analytics/metrics-authorization-test", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
	after := analyticsRequestCount(t, "/analytics/metrics-authorization-test", "401")
	if after != before+1 {
		t.Fatalf("expected 401 metric count to increase by one, got %d before and %d after", before, after)
	}
}

func TestAnalyticsMetricsRecordsRejectedBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Init(true)
	router := InitializeRouter()
	router.GET("/analytics/metrics-bearer-test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	before := analyticsRequestCount(t, "/analytics/metrics-bearer-test", "401")
	request := httptest.NewRequest(http.MethodGet, "/analytics/metrics-bearer-test", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
	after := analyticsRequestCount(t, "/analytics/metrics-bearer-test", "401")
	if after != before+1 {
		t.Fatalf("expected 401 metric count to increase by one, got %d before and %d after", before, after)
	}
}

func analyticsRequestCount(t *testing.T, route string, status string) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != "sentinel_analytics_request_duration_seconds" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["route"] == route && labels["status"] == status {
				return metric.GetHistogram().GetSampleCount()
			}
		}
	}
	return 0
}
