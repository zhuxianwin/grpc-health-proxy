package health

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func bufLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestAudit_LogsOnCheck(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewAuditChecker(inner, AuditConfig{Logger: bufLogger(&buf)})

	c.Check(context.Background(), "svc")

	if !strings.Contains(buf.String(), "health_check") {
		t.Fatalf("expected health_check log, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "svc") {
		t.Fatalf("expected service name in log, got: %s", buf.String())
	}
}

func TestAudit_LogsError(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("boom")}}
	c := NewAuditChecker(inner, AuditConfig{Logger: bufLogger(&buf)})

	c.Check(context.Background(), "svc")

	if !strings.Contains(buf.String(), "boom") {
		t.Fatalf("expected error in log, got: %s", buf.String())
	}
}

func TestAudit_PassesThroughResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewAuditChecker(inner, DefaultAuditConfig())

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestAudit_NilLoggerUsesDefault(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	// should not panic
	c := NewAuditChecker(inner, AuditConfig{})
	c.Check(context.Background(), "svc")
}

func TestAudit_CustomFields(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := AuditConfig{
		Logger:       bufLogger(&buf),
		ServiceField: "svc_name",
		LatencyField: "dur_ms",
	}
	c := NewAuditChecker(inner, cfg)
	c.Check(context.Background(), "mysvc")

	if !strings.Contains(buf.String(), "svc_name") {
		t.Fatalf("expected custom service field, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "dur_ms") {
		t.Fatalf("expected custom latency field, got: %s", buf.String())
	}
}
