package health

import (
	"context"
	"log/slog"
	"time"
)

// WarmupConfig controls how the warmup checker behaves.
type WarmupConfig struct {
	// Duration is how long to suppress unhealthy results after startup.
	Duration time.Duration
	// MinChecks is the minimum number of successful checks before reporting healthy.
	MinChecks int
}

// DefaultWarmupConfig returns sensible defaults.
func DefaultWarmupConfig() WarmupConfig {
	return WarmupConfig{
		Duration:  10 * time.Second,
		MinChecks: 1,
	}
}

// warmupChecker suppresses unhealthy results during the warmup window.
type warmupChecker struct {
	inner     Checker
	cfg       WarmupConfig
	start     time.Time
	successes int
	log       *slog.Logger
}

// NewWarmupChecker wraps inner and suppresses unhealthy results until the
// warmup window has elapsed and MinChecks successful checks have been seen.
func NewWarmupChecker(inner Checker, cfg WarmupConfig, log *slog.Logger) Checker {
	if log == nil {
		log = slog.Default()
	}
	return &warmupChecker{inner: inner, cfg: cfg, start: time.Now(), log: log}
}

func (w *warmupChecker) Check(ctx context.Context, service string) Result {
	res := w.inner.Check(ctx, service)

	if res.Status == StatusHealthy {
		w.successes++
	}

	warm := time.Since(w.start) >= w.cfg.Duration && w.successes >= w.cfg.MinChecks
	if !warm && res.Status != StatusHealthy {
		w.log.Info("warmup: suppressing unhealthy result",
			"service", service,
			"elapsed", time.Since(w.start).Round(time.Millisecond),
			"successes", w.successes,
		)
		return Result{Status: StatusHealthy, Message: "warming up"}
	}
	return res
}
