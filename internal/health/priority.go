package health

import (
	"context"
	"fmt"
	"log/slog"
)

// PriorityConfig controls how the priority checker selects among multiple
// checkers ordered by precedence. The first checker that returns a non-error
// result (healthy or explicitly unhealthy) wins; subsequent checkers are only
// consulted when higher-priority ones return an error.
type PriorityConfig struct {
	// Checkers is an ordered slice of (name, Checker) pairs, highest priority first.
	Checkers []PriorityEntry
}

// PriorityEntry pairs a human-readable name with a Checker.
type PriorityEntry struct {
	Name    string
	Checker Checker
}

type priorityChecker struct {
	cfg    PriorityConfig
	logger *slog.Logger
}

// NewPriorityChecker returns a Checker that tries each entry in order and
// returns the first conclusive result. If all checkers fail with errors the
// last error is returned.
func NewPriorityChecker(cfg PriorityConfig, logger *slog.Logger) Checker {
	if logger == nil {
		logger = slog.Default()
	}
	if len(cfg.Checkers) == 0 {
		panic("priority: at least one checker is required")
	}
	return &priorityChecker{cfg: cfg, logger: logger}
}

func (p *priorityChecker) Check(ctx context.Context, service string) Result {
	var lastErr error

	for _, entry := range p.cfg.Checkers {
		result := entry.Checker.Check(ctx, service)
		if result.Err == nil {
			// Conclusive answer (healthy or unhealthy with no transport error).
			p.logger.Debug("priority checker resolved",
				"checker", entry.Name,
				"service", service,
				"status", result.Status.String(),
			)
			return result
		}
		p.logger.Warn("priority checker skipping to next",
			"checker", entry.Name,
			"service", service,
			"err", result.Err,
		)
		lastErr = result.Err
	}

	return Result{
		Status: StatusUnknown,
		Err:    fmt.Errorf("all priority checkers failed; last error: %w", lastErr),
	}
}
