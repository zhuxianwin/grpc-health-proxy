package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// ShadowConfig controls shadow-check behaviour.
type ShadowConfig struct {
	// SampleRate is the fraction of requests that trigger a shadow check [0.0, 1.0].
	SampleRate float64
	// Timeout caps the shadow goroutine.
	Timeout time.Duration
}

// DefaultShadowConfig returns sensible defaults.
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		SampleRate: 0.1,
		Timeout:    2 * time.Second,
	}
}

// shadowChecker forwards every call to primary and, with probability
// SampleRate, concurrently calls shadow without affecting the result.
type shadowChecker struct {
	primary    Checker
	shadow     Checker
	cfg        ShadowConfig
	log        *slog.Logger
	mu         sync.Mutex
	counter    uint64
	sampleEvery uint64
}

// NewShadowChecker wraps primary and fires shadow checks on a sampled subset
// of requests. Shadow results are logged but never returned to the caller.
func NewShadowChecker(primary, shadow Checker, cfg ShadowConfig, log *slog.Logger) Checker {
	if log == nil {
		log = slog.Default()
	}
	rate := cfg.SampleRate
	if rate <= 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	every := uint64(0)
	if rate > 0 {
		every = uint64(1.0 / rate)
		if every == 0 {
			every = 1
		}
	}
	return &shadowChecker{
		primary:     primary,
		shadow:      shadow,
		cfg:         cfg,
		log:         log,
		sampleEvery: every,
	}
}

func (s *shadowChecker) Check(ctx context.Context, service string) Result {
	result := s.primary.Check(ctx, service)

	if s.sampleEvery == 0 {
		return result
	}
	s.mu.Lock()
	s.counter++
	fire := s.counter%s.sampleEvery == 0
	s.mu.Unlock()

	if fire {
		go func() {
			tctx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
			defer cancel()
			sr := s.shadow.Check(tctx, service)
			s.log.Info("shadow check result",
				"service", service,
				"status", sr.Status.String(),
				"err", sr.Err,
			)
		}()
	}
	return result
}
