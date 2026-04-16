package health

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

// QuotaConfig controls periodic quota reset behaviour.
type QuotaConfig struct {
	// MaxCalls is the maximum number of health checks allowed per Window.
	MaxCalls int64
	// Window is the duration after which the counter resets.
	Window time.Duration
}

// DefaultQuotaConfig returns sensible defaults.
func DefaultQuotaConfig() QuotaConfig {
	return QuotaConfig{
		MaxCalls: 60,
		Window:   time.Minute,
	}
}

// quotaChecker wraps a Checker and enforces a call quota per time window.
type quotaChecker struct {
	inner  Checker
	cfg    QuotaConfig
	count  atomic.Int64
	logger *slog.Logger
}

// NewQuotaChecker returns a Checker that rejects calls once the quota is
// exhausted, resetting the counter every cfg.Window.
func NewQuotaChecker(inner Checker, cfg QuotaConfig, logger *slog.Logger) Checker {
	if cfg.MaxCalls <= 0 {
		cfg.MaxCalls = DefaultQuotaConfig().MaxCalls
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultQuotaConfig().Window
	}
	if logger == nil {
		logger = slog.Default()
	}
	qc := &quotaChecker{inner: inner, cfg: cfg, logger: logger}
	go qc.resetLoop()
	return qc
}

func (q *quotaChecker) resetLoop() {
	ticker := time.NewTicker(q.cfg.Window)
	defer ticker.Stop()
	for range ticker.C {
		q.count.Store(0)
	}
}

func (q *quotaChecker) Check(ctx context.Context, service string) Result {
	n := q.count.Add(1)
	if n > q.cfg.MaxCalls {
		q.logger.Warn("quota exceeded", "service", service, "count", n, "max", q.cfg.MaxCalls)
		return Result{Err: fmt.Errorf("quota exceeded (%d/%d)", n, q.cfg.MaxCalls)}
	}
	return q.inner.Check(ctx, service)
}
