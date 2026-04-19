package health

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func newBufLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), buf
}

func TestTrace_DelegatesResult(t *testing.T) {
	inner := staticChecker{Result{Status: StatusHealthy}}
	logger, _ := newBufLogger()
	c := NewTraceChecker(inner, DefaultTraceConfig(logger))
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestTrace_LogsOnCheck(t *testing.T) {
	inner := staticChecker{Result{Status: StatusHealthy}}
	logger, buf := newBufLogger()
	c := NewTraceChecker(inner, DefaultTraceConfig(logger))
	c.Check(context.Background(), "my-service")
	if !bytes.Contains(buf.Bytes(), []byte("my-service")) {
		t.Fatalf("expected service name in log output, got: %s", buf.String())
	}
}

func TestTrace_LogsError(t *testing.T) {
	inner := staticChecker{Result{Status: StatusUnhealthy, Err: errors.New("boom")}}
	logger, buf := newBufLogger()
	c := NewTraceChecker(inner, DefaultTraceConfig(logger))
	c.Check(context.Background(), "svc")
	if !bytes.Contains(buf.Bytes(), []byte("boom")) {
		t.Fatalf("expected error in log output, got: %s", buf.String())
	}
}

func TestTrace_NilLoggerUsesDefault(t *testing.T) {
	inner := staticChecker{Result{Status: StatusHealthy}}
	c := NewTraceChecker(inner, TraceConfig{Logger: nil})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}

func TestTraceHandler_ReturnsJSON(t *testing.T) {
	logger, _ := newBufLogger()
	inner := staticChecker{Result{Status: StatusHealthy}}
	c := NewTraceChecker(inner, DefaultTraceConfig(logger))
	h := NewTraceStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", resp["enabled"])
	}
}

func TestTraceHandler_NotATraceChecker(t *testing.T) {
	inner := staticChecker{Result{Status: StatusHealthy}}
	h := NewTraceStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["enabled"] != false {
		t.Fatalf("expected enabled=false, got %v", resp["enabled"])
	}
}
