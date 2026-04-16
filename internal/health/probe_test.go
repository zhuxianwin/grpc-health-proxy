package health

import (
	"context"
	"errors"
	"net//http/httptest"
	"testing"
	"time"
)

type stubChecker struct {
	result Result (s *stubChecker) Check(_ context.Context, _ string) Result { return s.result }

func TestProbeHandler_Healthy(t *testing.T) ecker := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewProbeHandler(checker, DefaultProbeConfig(), nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProbeHandler_Unhealthy(t *testing.T) {
	checker := &stubChecker{result: Result{Status: StatusUnhealthy}}
	h := NewProbeHandler(checker, DefaultProbeConfig(), nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestProbeHandler_Error(t *testing.T) {
	checker := &stubChecker{result: Result{Status: StatusUnknown, Err: errors.New("dial failed")}}
	h := NewProbeHandler(checker, DefaultProbeConfig(), nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestProbeHandler_DefaultTimeoutOnZero(t *testing.T) {
	checker := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := ProbeConfig{Timeout: 0, ServiceName: "my.Service"}
	h := NewProbeHandler(checker, cfg, nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDefaultProbeConfig(t *testing.T) {
	cfg := DefaultProbeConfig()
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("unexpected default timeout: %v", cfg.Timeout)
	}
	if cfg.ServiceName != "" {
		t.Fatalf("expected empty service name, got %q", cfg.ServiceName)
	}
}
