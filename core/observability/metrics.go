package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var analyticsRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "sentinel",
		Subsystem: "analytics",
		Name:      "request_duration_seconds",
		Help:      "Duration of analytics HTTP requests.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"route", "status"},
)

var databaseQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "sentinel",
		Subsystem: "database",
		Name:      "query_duration_seconds",
		Help:      "Duration of database queries by operation and result.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
	},
	[]string{"operation", "result"},
)

var databaseSlowQueries = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "sentinel",
		Subsystem: "database",
		Name:      "slow_queries_total",
		Help:      "Database queries exceeding the configured slow-query threshold.",
	},
	[]string{"operation"},
)

func init() {
	prometheus.MustRegister(analyticsRequestDuration, databaseQueryDuration, databaseSlowQueries)
}

func ObserveAnalyticsRequest(route string, status int, elapsed time.Duration) {
	if route == "" {
		route = "unknown"
	}
	analyticsRequestDuration.WithLabelValues(route, strconv.Itoa(status)).Observe(elapsed.Seconds())
}

func ObserveDatabaseQuery(operation string, failed bool, slow bool, elapsed time.Duration) {
	result := "success"
	if failed {
		result = "error"
	}
	databaseQueryDuration.WithLabelValues(operation, result).Observe(elapsed.Seconds())
	if slow {
		databaseSlowQueries.WithLabelValues(operation).Inc()
	}
}
