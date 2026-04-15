package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimitConfig controls how frequently health checks may be performed.
type RateLimitConfig struct {
	// MaxChecksPerSecond is the maximum number of check calls allowed per second.
	MaxChecksPerSecond float64
	// Burst is the maximum burst size above the steady-state rate.
	Burst int
}

// DefaultRateLimitConfig returns a sensible default rate-limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxChecksPerSecond: 10,
		Burst:              5,
	}
}

// tokenBucket is a minimal token-bucket rate limiter.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per nanosecond
	lastTick time.Time
}

func newTokenBucket(rps float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:   float64(burst),
		max:      float64(burst),
		rate:     rps / float64(time.Second),
		lastTick: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(tb.lastTick)
	tb.lastTick = now
	tb.tokens += float64(elapsed) * tb.rate
	if tb.tokens > tb.max {
		tb.tokens = tb.max
	}
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimitedChecker wraps a Checker and enforces a token-bucket rate limit.
type RateLimitedChecker struct {
	inner  Checker
	bucket *tokenBucket
	service string
}

// NewRateLimitedChecker wraps inner with the supplied rate-limit configuration.
func NewRateLimitedChecker(inner Checker, service string, cfg RateLimitConfig) *RateLimitedChecker {
	return &RateLimitedChecker{
		inner:   inner,
		bucket:  newTokenBucket(cfg.MaxChecksPerSecond, cfg.Burst),
		service: service,
	}
}

// Check performs the health check if a token is available; otherwise it
// returns an error result indicating the rate limit was exceeded.
func (r *RateLimitedChecker) Check(ctx context.Context) Result {
	if !r.bucket.allow() {
		return Result{
			Status: StatusUnknown,
			Err:    fmt.Errorf("rate limit exceeded for service %q", r.service),
		}
	}
	return r.inner.Check(ctx)
}
