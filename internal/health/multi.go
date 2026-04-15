package health

import (
	"context"
	"sync"
)

// MultiChecker runs health checks against multiple services concurrently
// and aggregates the results.
type MultiChecker struct {
	checkers map[string]*Checker
	cache    *Cache
}

// NewMultiChecker creates a MultiChecker from a list of target addresses.
func NewMultiChecker(targets []string, cache *Cache) *MultiChecker {
	checkers := make(map[string]*Checker, len(targets))
	for _, t := range targets {
		checkers[t] = NewChecker(t, cache)
	}
	return &MultiChecker{checkers: checkers, cache: cache}
}

// CheckAll runs all checks concurrently and returns a map of target -> Result.
func (m *MultiChecker) CheckAll(ctx context.Context, service string) map[string]Result {
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make(map[string]Result, len(m.checkers))

	for target, c := range m.checkers {
		wg.Add(1)
		go func(t string, chk *Checker) {
			defer wg.Done()
			r := chk.Check(ctx, service)
			mu.Lock()
			results[t] = r
			mu.Unlock()
		}(target, c)
	}

	wg.Wait()
	return results
}

// Healthy returns true only when all targets report a healthy status.
func (m *MultiChecker) Healthy(results map[string]Result) bool {
	for _, r := range results {
		if r.Status != StatusHealthy {
			return false
		}
	}
	return true
}
