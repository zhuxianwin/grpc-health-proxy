package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultBudgetConfig returns a BudgetConfig with sensible defaults.
func DefaultBudgetConfig() BudgetConfig {
	return BudgetConfig{
		Window:       30 * time.Second,
		ErrorBudget:  0.10, // 10% error budget
		MinSamples:   5,
		Logger:       slog.Default(),
	}
}

// BudgetConfig controls the error-budget checker behaviour.
type BudgetConfig struct {
	Window      time.Duration
	ErrorBudget float64 // fraction of allowed errors, e.g. 0.10 = 10%
	MinSamples  int
	Logger      *slog.Logger
}

type budgetEntry struct {
	at  time.Time
	ok  bool
}

// BudgetChecker tracks an error budget over a sliding window. Once the
// fraction of failing checks exceeds ErrorBudget the checker returns
// Unhealthy until the budget recovers.
type BudgetChecker struct {
	inner  Checker
	cfg    BudgetConfig
	mu     sync.Mutex
	window map[string][]budgetEntry
}

// NewBudgetChecker wraps inner with error-budget enforcement.
func NewBudgetChecker(inner Checker, cfg BudgetConfig) *BudgetChecker {
	if cfg.Window == 0 {
		cfg = DefaultBudgetConfig()
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &BudgetChecker{inner: inner, cfg: cfg, window: make(map[string][]budgetEntry)}
}

// Check delegates to inner, records the outcome, and returns Unhealthy when
// the error budget is exhausted.
func (b *BudgetChecker) Check(ctx context.Context, service string) Result {
	res := b.inner.Check(ctx, service)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-b.cfg.Window)

	entries := b.window[service]
	// evict old entries
	filtered := entries[:0]
	for _, e := range entries {
		if e.at.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, budgetEntry{at: now, ok: res.Status == StatusHealthy})
	b.window[service] = filtered

	if len(filtered) < b.cfg.MinSamples {
		return res
	}

	var failures int
	for _, e := range filtered {
		if !e.ok {
			failures++
		}
	}
	errorRate := float64(failures) / float64(len(filtered))
	if errorRate > b.cfg.ErrorBudget {
		b.cfg.Logger.Warn("error budget exhausted",
			"service", service,
			"error_rate", errorRate,
			"budget", b.cfg.ErrorBudget,
		)
		return Result{Status: StatusUnhealthy, Err: ErrBudgetExhausted}
	}
	return res
}

// BudgetStats returns current window statistics for a service.
func (b *BudgetChecker) BudgetStats(service string) (total, failures int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entries := b.window[service]
	for _, e := range entries {
		total++
		if !e.ok {
			failures++
		}
	}
	return
}
