package health

import (
	"context"
	"log/slog"
	"time"
)

// HedgeConfig controls hedged-request behaviour.
type HedgeConfig struct {
	// Delay is how long to wait before issuing the second request.
	Delay time.Duration
	// MaxHedged is the maximum number of concurrent hedged requests (including the original).
	MaxHedged int
}

// DefaultHedgeConfig returns a sensible default.
func DefaultHedgeConfig() HedgeConfig {
	return HedgeConfig{
		Delay:     50 * time.Millisecond,
		MaxHedged: 2,
	}
}

type hedgeChecker struct {
	inner  Checker
	cfg    HedgeConfig
	logger *slog.Logger
}

// NewHedgeChecker wraps inner so that if no result arrives within cfg.Delay a
// second (hedged) request is fired concurrently. The first successful result
// wins; errors are returned only when all hedges fail.
func NewHedgeChecker(inner Checker, cfg HedgeConfig, logger *slog.Logger) Checker {
	if cfg.MaxHedged < 1 {
		cfg.MaxHedged = DefaultHedgeConfig().MaxHedged
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &hedgeChecker{inner: inner, cfg: cfg, logger: logger}
}

func (h *hedgeChecker) Check(ctx context.Context, service string) Result {
	type outcome struct {
		res Result
		idx int
	}

	max := h.cfg.MaxHedged
	results := make(chan outcome, max)

	launch := func(idx int) {
		res := h.inner.Check(ctx, service)
		results <- outcome{res, idx}
	}

	go launch(0)

	var lastErr Result
	received := 0
	nextHedge := 1
	timer := time.NewTimer(h.cfg.Delay)
	defer timer.Stop()

	for received < max {
		select {
		case o := <-results:
			received++
			if o.res.Err == nil {
				h.logger.Debug("hedge: request succeeded", "attempt", o.idx, "service", service)
				return o.res
			}
			lastErr = o.res
		case <-timer.C:
			if nextHedge < max {
				h.logger.Debug("hedge: firing hedged request", "attempt", nextHedge, "service", service)
				go launch(nextHedge)
				nextHedge++
				timer.Reset(h.cfg.Delay)
			}
		case <-ctx.Done():
			return Result{Status: StatusUnhealthy, Err: ctx.Err()}
		}
	}
	return lastErr
}
