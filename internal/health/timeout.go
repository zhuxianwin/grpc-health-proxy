package health

import (
	"context"
	"fmt"
	"time"
)

// TimeoutChecker wraps a Checker and enforces a per-check deadline.
type TimeoutChecker struct {
	inner   Checker
	timeout time.Duration
}

// NewTimeoutChecker returns a Checker that cancels the underlying call
// if it does not complete within the given timeout.
func NewTimeoutChecker(inner Checker, timeout time.Duration) *TimeoutChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &TimeoutChecker{inner: inner, timeout: timeout}
}

// Check delegates to the inner Checker with a bounded context.
func (t *TimeoutChecker) Check(ctx context.Context, service string) Result {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type outcome struct {
		res Result
	}
	ch := make(chan outcome, 1)

	go func() {
		ch <- outcome{res: t.inner.Check(ctx, service)}
	}()

	select {
	case o := <-ch:
		return o.res
	case <-ctx.Done():
		return Result{
			Service: service,
			Status:  StatusUnknown,
			Err:     fmt.Errorf("health check timed out after %s", t.timeout),
		}
	}
}
