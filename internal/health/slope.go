package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultSlopeConfig returns a SlopeConfig with sensible defaults.
func DefaultSlopeConfig() SlopeConfig {
	return SlopeConfig{
		Window:    10,
		Threshold: -0.3,
	}
}

// SlopeConfig controls the slope checker behaviour.
type SlopeConfig struct {
	// Window is the number of recent results to consider.
	Window int
	// Threshold is the minimum acceptable slope (rise/run over window).
	// A negative value means the healthy-rate is declining; values below
	// this threshold cause the checker to return Unhealthy.
	Threshold float64
}

type slopeChecker struct {
	inner  Checker
	cfg    SlopeConfig
	log    *slog.Logger
	mu     sync.Mutex
	bucket map[string][]bool
}

// NewSlopeChecker wraps inner and returns Unhealthy when the slope of the
// healthy-rate over the configured window drops below cfg.Threshold.
func NewSlopeChecker(inner Checker, cfg SlopeConfig, log *slog.Logger) Checker {
	if cfg.Window <= 0 {
		cfg = DefaultSlopeConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &slopeChecker{
		inner:  inner,
		cfg:    cfg,
		log:    log,
		bucket: make(map[string][]bool),
	}
}

func (s *slopeChecker) Check(ctx context.Context, service string) Result {
	res := s.inner.Check(ctx, service)

	s.mu.Lock()
	buf := append(s.bucket[service], res.Status == StatusHealthy)
	if len(buf) > s.cfg.Window {
		buf = buf[len(buf)-s.cfg.Window:]
	}
	s.bucket[service] = buf
	slope := computeSlope(buf)
	s.mu.Unlock()

	if len(buf) >= s.cfg.Window && slope < s.cfg.Threshold {
		s.log.Warn("slope checker: declining healthy-rate",
			"service", service, "slope", slope, "threshold", s.cfg.Threshold)
		return Result{
			Status:    StatusUnhealthy,
			CheckedAt: time.Now(),
			Err:       nil,
		}
	}
	return res
}

// computeSlope returns the least-squares slope of a boolean series treated
// as 1.0 (healthy) or 0.0 (unhealthy).
func computeSlope(pts []bool) float64 {
	n := float64(len(pts))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range pts {
		x := float64(i)
		var y float64
		if v {
			y = 1
		}
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
