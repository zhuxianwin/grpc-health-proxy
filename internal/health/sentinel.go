package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultSentinelConfig returns a SentinelConfig with sensible defaults.
func DefaultSentinelConfig() SentinelConfig {
	return SentinelConfig{
		TripAfter:   3,
		ResetAfter:  30 * time.Second,
		HealthyCode: StatusHealthy,
	}
}

// SentinelConfig controls how the sentinel checker behaves.
type SentinelConfig struct {
	// TripAfter is the number of consecutive unhealthy results before the sentinel fires.
	TripAfter int
	// ResetAfter is the duration after tripping before the sentinel resets.
	ResetAfter time.Duration
	// HealthyCode is the status returned while the sentinel is tripped.
	HealthyCode Status
	// Logger is an optional structured logger.
	Logger *slog.Logger
}

type sentinelChecker struct {
	cfg       SentinelConfig
	inner     Checker
	mu        sync.Mutex
	consec    map[string]int
	trippedAt map[string]time.Time
	log       *slog.Logger
}

// NewSentinelChecker wraps inner and trips after cfg.TripAfter consecutive
// unhealthy results, returning a synthetic healthy result until cfg.ResetAfter
// elapses. This is useful for absorbing transient cascading failures.
func NewSentinelChecker(inner Checker, cfg SentinelConfig) Checker {
	if cfg.TripAfter <= 0 {
		cfg = DefaultSentinelConfig()
	}
	if cfg.ResetAfter <= 0 {
		cfg.ResetAfter = DefaultSentinelConfig().ResetAfter
	}
	l := cfg.Logger
	if l == nil {
		l = slog.Default()
	}
	return &sentinelChecker{
		cfg:       cfg,
		inner:     inner,
		consec:    make(map[string]int),
		trippedAt: make(map[string]time.Time),
		log:       l,
	}
}

func (s *sentinelChecker) Check(ctx context.Context, service string) Result {
	s.mu.Lock()
	if t, ok := s.trippedAt[service]; ok {
		if time.Since(t) >= s.cfg.ResetAfter {
			delete(s.trippedAt, service)
			delete(s.consec, service)
			s.log.Info("sentinel reset", "service", service)
		} else {
			s.mu.Unlock()
			s.log.Debug("sentinel tripped, suppressing", "service", service)
			return Result{Service: service, Status: s.cfg.HealthyCode}
		}
	}
	s.mu.Unlock()

	res := s.inner.Check(ctx, service)

	s.mu.Lock()
	defer s.mu.Unlock()
	if res.Status != StatusHealthy || res.Err != nil {
		s.consec[service]++
		if s.consec[service] >= s.cfg.TripAfter {
			s.trippedAt[service] = time.Now()
			s.log.Warn("sentinel tripped", "service", service, "consecutive", s.consec[service])
		}
	} else {
		s.consec[service] = 0
	}
	return res
}
