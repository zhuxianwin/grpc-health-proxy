package health

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// DefaultSheddingConfig returns a conservative load-shedding configuration.
func DefaultSheddingConfig() SheddingConfig {
	return SheddingConfig{
		MaxLatency:   500 * time.Millisecond,
		WindowSize:   10,
		SheddingRatio: 0.8,
	}
}

// SheddingConfig controls when the load-shedding checker starts rejecting calls.
type SheddingConfig struct {
	// MaxLatency is the p-latency threshold above which load shedding activates.
	MaxLatency time.Duration
	// WindowSize is the number of recent observations to track.
	WindowSize int
	// SheddingRatio is the fraction of WindowSize that must exceed MaxLatency
	// before requests are shed (0 < ratio <= 1).
	SheddingRatio float64
}

// sheddingChecker wraps a Checker and sheds load when recent latencies are too high.
type sheddingChecker struct {
	inner  Checker
	cfg    SheddingConfig
	log    *slog.Logger

	// circular latency window stored as nanoseconds
	window []int64
	pos    atomic.Int64 // next write index (mod WindowSize)
	count  atomic.Int64 // total writes, capped at WindowSize for ratio calc
}

// NewSheddingChecker wraps inner with latency-based load shedding.
func NewSheddingChecker(inner Checker, cfg SheddingConfig, log *slog.Logger) Checker {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = DefaultSheddingConfig().WindowSize
	}
	if cfg.SheddingRatio <= 0 || cfg.SheddingRatio > 1 {
		cfg.SheddingRatio = DefaultSheddingConfig().SheddingRatio
	}
	if log == nil {
		log = slog.Default()
	}
	return &sheddingChecker{
		inner:  inner,
		cfg:    cfg,
		log:    log,
		window: make([]int64, cfg.WindowSize),
	}
}

func (s *sheddingChecker) Check(ctx context.Context, service string) Result {
	if s.shouldShed() {
		s.log.Warn("load shedding active, rejecting health check", "service", service)
		return Result{Status: StatusUnhealthy, Err: ErrLoadShed}
	}

	start := time.Now()
	res := s.inner.Check(ctx, service)
	s.record(time.Since(start))
	return res
}

func (s *sheddingChecker) record(d time.Duration) {
	idx := int(s.pos.Add(1)-1) % s.cfg.WindowSize
	s.window[idx] = d.Nanoseconds()
	if s.count.Load() < int64(s.cfg.WindowSize) {
		s.count.Add(1)
	}
}

func (s *sheddingChecker) shouldShed() bool {
	n := int(s.count.Load())
	if n == 0 {
		return false
	}
	threshold := s.cfg.MaxLatency.Nanoseconds()
	exceeded := 0
	for i := 0; i < n; i++ {
		if s.window[i] > threshold {
			exceeded++
		}
	}
	return float64(exceeded)/float64(n) >= s.cfg.SheddingRatio
}
