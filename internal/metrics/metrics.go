// Package metrics provides Prometheus metrics for the gRPC health proxy.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HealthCheckTotal counts the total number of health check requests.
	HealthCheckTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grpc_health_proxy",
			Name:      "health_checks_total",
			Help:      "Total number of gRPC health check requests made.",
		},
		[]string{"service", "status"},
	)

	// HealthCheckDuration observes the duration of health check requests.
	HealthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "grpc_health_proxy",
			Name:      "health_check_duration_seconds",
			Help:      "Duration of gRPC health check requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service"},
	)

	// HTTPRequestsTotal counts HTTP requests to the proxy's own endpoints.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grpc_health_proxy",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests received by the proxy.",
		},
		[]string{"path", "code"},
	)
)

// Handler returns an HTTP handler that serves Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
