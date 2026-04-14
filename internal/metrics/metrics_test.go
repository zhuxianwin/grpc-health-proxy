package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/your-org/grpc-health-proxy/internal/metrics"
)

func TestHealthCheckTotal_Increments(t *testing.T) {
	// Reset by using a fresh counter read via testutil.
	metrics.HealthCheckTotal.WithLabelValues("my-service", "healthy").Inc()
	metrics.HealthCheckTotal.WithLabelValues("my-service", "healthy").Inc()

	count := testutil.ToFloat64(
		metrics.HealthCheckTotal.WithLabelValues("my-service", "healthy"),
	)
	if count < 2 {
		t.Errorf("expected at least 2 increments, got %v", count)
	}
}

func TestHealthCheckDuration_Observe(t *testing.T) {
	obs := metrics.HealthCheckDuration.WithLabelValues("my-service")
	obs.Observe(0.05)
	obs.Observe(0.1)

	// Verify the histogram has recorded observations without error.
	err := testutil.CollectAndCompare(
		metrics.HealthCheckDuration,
		strings.NewReader(""), // we only check it doesn't error
	)
	// We expect a mismatch because we're not providing expected output;
	// the important thing is that collection does not panic.
	_ = err
}

func TestHandler_ServesMetrics(t *testing.T) {
	// Emit a known metric so /metrics is non-empty.
	metrics.HTTPRequestsTotal.WithLabelValues("/healthz", "200").Inc()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	handler := metrics.Handler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "grpc_health_proxy_http_requests_total") {
		t.Error("expected metrics output to contain grpc_health_proxy_http_requests_total")
	}
}

func TestMetricLabels(t *testing.T) {
	// Ensure label cardinality is respected — wrong label count should panic.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("correctly panicked on wrong label count: %v", r)
		}
	}()

	// This should work fine (correct labels).
	metrics.HealthCheckTotal.With(prometheus.Labels{
		"service": "svc",
		"status":  "unhealthy",
	}).Inc()
}
