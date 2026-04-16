package health

import (
	"context"
	"sync"

	"golang.org/x/sync/singleflight"
)

// DedupChecker wraps a Checker and collapses concurrent in-flight requests
// for the same service into a single upstream call. This prevents thundering-
// herd problems when many probes arrive simultaneously.
type DedupChecker struct {
	inner   Checker
	group   singleflight.Group
	mu      sync.Mutex
}

// NewDedupChecker returns a DedupChecker that deduplicates concurrent Check
// calls that share the same service name.
func NewDedupChecker(inner Checker) *DedupChecker {
	if inner == nil {
		panic("dedup: inner checker must not be nil")
	}
	return &DedupChecker{inner: inner}
}

// Check performs a health check, merging any concurrent calls for the same
// service into a single in-flight request.
func (d *DedupChecker) Check(ctx context.Context, service string) Result {
	type payload struct {
		result Result
	}

	v, _, _ := d.group.Do(service, func() (interface{}, error) {
		r := d.inner.Check(ctx, service)
		return payload{result: r}, nil
	})

	return v.(payload).result
}

// Forget removes the in-flight record for service, allowing the next call to
// start a fresh request. Useful in tests or after a known topology change.
func (d *DedupChecker) Forget(service string) {
	d.group.Forget(service)
}
