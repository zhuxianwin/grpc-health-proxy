package health

import (
	"context"
	"sync"
)

// BatchConfig controls parallel fan-out behaviour.
type BatchConfig struct {
	// MaxConcurrency limits simultaneous checks. 0 means unlimited.
	MaxConcurrency int
}

// DefaultBatchConfig returns sensible defaults.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{MaxConcurrency: 8}
}

// BatchResult pairs a target name with its Result.
type BatchResult struct {
	Target string
	Result Result
}

// BatchChecker runs multiple Checkers concurrently and returns all results.
type BatchChecker struct {
	targets []string
	checkers map[string]Checker
	sem      chan struct{}
}

// NewBatchChecker creates a BatchChecker for the given target→checker map.
func NewBatchChecker(checkers map[string]Checker, cfg BatchConfig) *BatchChecker {
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = DefaultBatchConfig().MaxConcurrency
	}
	targets := make([]string, 0, len(checkers))
	for t := range checkers {
		targets = append(targets, t)
	}
	sem := make(chan struct{}, cfg.MaxConcurrency)
	return &BatchChecker{targets: targets, checkers: checkers, sem: sem}
}

// CheckAll runs all checkers concurrently and returns a slice of BatchResults.
func (b *BatchChecker) CheckAll(ctx context.Context) []BatchResult {
	results := make([]BatchResult, len(b.targets))
	var wg sync.WaitGroup
	for i, t := range b.targets {
		wg.Add(1)
		go func(idx int, target string) {
			defer wg.Done()
			select {
			case b.sem <- struct{}{}:
				defer func() { <-b.sem }()
			case <-ctx.Done():
				results[idx] = BatchResult{Target: target, Result: Result{Status: StatusUnknown, Err: ctx.Err()}}
				return
			}
			r := b.checkers[target].Check(ctx, target)
			results[idx] = BatchResult{Target: target, Result: r}
		}(i, t)
	}
	wg.Wait()
	return results
}
