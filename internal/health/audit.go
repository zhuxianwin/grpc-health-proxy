package health

import (
	"context"
	"log/slog"
	"time"
)

// AuditConfig controls audit logging behaviour.
type AuditConfig struct {
	// Logger receives one structured log line per health check.
	Logger *slog.Logger
	// ServiceField is the log attribute key for the service name (default: "service").
	ServiceField string
	// LatencyField is the log attribute key for elapsed time (default: "latency_ms").
	LatencyField string
}

// DefaultAuditConfig returns sensible defaults.
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Logger:       slog.Default(),
		ServiceField: "service",
		LatencyField: "latency_ms",
	}
}

type auditChecker struct {
	inner  Checker
	cfg    AuditConfig
}

// NewAuditChecker wraps inner and emits a structured log line for every check.
func NewAuditChecker(inner Checker, cfg AuditConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.ServiceField == "" {
		cfg.ServiceField = "service"
	}
	if cfg.LatencyField == "" {
		cfg.LatencyField = "latency_ms"
	}
	return &auditChecker{inner: inner, cfg: cfg}
}

func (a *auditChecker) Check(ctx context.Context, service string) Result {
	start := time.Now()
	res := a.inner.Check(ctx, service)
	elapsed := time.Since(start).Milliseconds()

	attrs := []any{
		slog.String(a.cfg.ServiceField, service),
		slog.Int64(a.cfg.LatencyField, elapsed),
		slog.String("status", res.Status.String()),
	}
	if res.Err != nil {
		attrs = append(attrs, slog.String("error", res.Err.Error()))
	}
	a.cfg.Logger.InfoContext(ctx, "health_check", attrs...)
	return res
}
