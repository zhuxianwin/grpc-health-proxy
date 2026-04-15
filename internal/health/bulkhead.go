package health

import (
	"context"
	"errors"
	"sync"
)

// ErrBulkheadFull is returned when the bulkhead concurrency limit is reached.
var ErrBulkheadFull = errors.New("bulkhead: max concurrent checks reached")

// BulkheadConfig holds configuration for the bulkhead limiter.
type BulkheadConfig struct {
	// MaxConcurrent is the maximum number of in-flight health checks allowed.
	MaxConcurrent int
}

// DefaultBulkheadConfig returns a BulkheadConfig with sensible defaults.
func DefaultBulkheadConfig() BulkheadConfig {
	return BulkheadConfig{
		MaxConcurrent: 10,
	}
}

// bulkheadChecker wraps a Checker and limits concurrent executions.
type bulkheadChecker struct {
	inner   Checker
	sem     chan struct{}
	mu      sync.Mutex
	inflight int
}

// NewBulkheadChecker wraps inner with a concurrency limiter. Calls that would
// exceed MaxConcurrent return ErrBulkheadFull immediately without blocking.
func NewBulkheadChecker(inner Checker, cfg BulkheadConfig) Checker {
	max := cfg.MaxConcurrent
	if max <= 0 {
		max = DefaultBulkheadConfig().MaxConcurrent
	}
	return &bulkheadChecker{
		inner: inner,
		sem:   make(chan struct{}, max),
	}
}

func (b *bulkheadChecker) Check(ctx context.Context, service string) Result {
	select {
	case b.sem <- struct{}{}:
		// acquired a slot
	default:
		return Result{
			Service: service,
			Status:  StatusUnhealthy,
			Err:     ErrBulkheadFull,
		}
	}
	defer func() { <-b.sem }()
	return b.inner.Check(ctx, service)
}

// Inflight returns the number of currently executing checks.
func (b *bulkheadChecker) Inflight() int {
	return len(b.sem)
}
