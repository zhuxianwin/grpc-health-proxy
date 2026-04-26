package health

import (
	"context"
	"log"
	"sync"
	"time"
)

// DefaultFloorConfig returns a FloorConfig with sensible defaults.
func DefaultFloorConfig() FloorConfig {
	return FloorConfig{
		MinStatus:    StatusHealthy,
		Window:       30 * time.Second,
		MinSamples:   3,
	}
}

// FloorConfig controls the behaviour of NewFloorChecker.
type FloorConfig struct {
	// MinStatus is the lowest status that will be returned; any result
	// below this floor is raised to MinStatus.
	MinStatus Status

	// Window is the observation window used to decide whether the floor
	// has been earned.  Results older than Window are discarded.
	Window time.Duration

	// MinSamples is the minimum number of samples required inside the
	// window before the floor is applied.  Until this threshold is met
	// the inner result is returned as-is.
	MinSamples int

	// Logger is optional.  When nil a default logger is used.
	Logger *log.Logger
}

type floorEntry struct {
	status Status
	at     time.Time
}

// FloorChecker wraps an inner Checker and ensures that once a service
// has demonstrated sustained health (MinSamples healthy results inside
// Window) its reported status never drops below MinStatus.
//
// This is useful when a downstream service is known to be flaky at
// startup and you want Kubernetes probes to remain green after an
// initial warm-up burst of healthy results.
type FloorChecker struct {
	inner  Checker
	cfg    FloorConfig
	mu     sync.Mutex
	window map[string][]floorEntry
}

// NewFloorChecker creates a FloorChecker wrapping inner.
// If cfg is zero-valued DefaultFloorConfig is used.
func NewFloorChecker(inner Checker, cfg FloorConfig, logger *log.Logger) *FloorChecker {
	if cfg.Window == 0 {
		cfg = DefaultFloorConfig()
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = DefaultFloorConfig().MinSamples
	}
	if logger != nil {
		cfg.Logger = logger
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &FloorChecker{
		inner:  inner,
		cfg:    cfg,
		window: make(map[string][]floorEntry),
	}
}

// Check delegates to the inner Checker, records the result in the
// sliding window, and raises the status to MinStatus when the floor
// condition is satisfied.
func (f *FloorChecker) Check(ctx context.Context, service string) Result {
	res := f.inner.Check(ctx, service)

	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-f.cfg.Window)

	// Append current observation and evict stale entries.
	entries := append(f.window[service], floorEntry{status: res.Status, at: now})
	valid := entries[:0]
	for _, e := range entries {
		if e.at.After(cutoff) {
			valid = append(valid, e)
		}
	}
	f.window[service] = valid

	// Only apply the floor once we have enough samples.
	if len(valid) < f.cfg.MinSamples {
		return res
	}

	// Count how many samples are at or above the floor.
	healthy := 0
	for _, e := range valid {
		if e.status >= f.cfg.MinStatus {
			healthy++
		}
	}

	// If all samples in the window meet the floor, enforce it.
	if healthy == len(valid) && res.Status < f.cfg.MinStatus {
		f.cfg.Logger.Printf("floor: raising %s status from %s to %s",
			service, res.Status, f.cfg.MinStatus)
		res.Status = f.cfg.MinStatus
	}

	return res
}
