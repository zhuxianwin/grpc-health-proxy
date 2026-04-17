// Package health — observe
//
// ObserveChecker wraps any Checker and invokes a user-supplied callback after
// every health check, reporting the service name, the Result, and the wall-clock
// duration of the inner call.
//
// Typical uses:
//   - Feeding latency and status data into custom metrics or dashboards.
//   - Logging slow checks without altering control flow.
//   - Integration testing — assert that a callback was called with expected values.
//
// Usage:
//
//	cfg := health.ObserveConfig{
//	    OnResult: func(svc string, r health.Result, d time.Duration) {
//	        myMetrics.Record(svc, r.Status.String(), d)
//	    },
//	}
//	checked := health.NewObserveChecker(inner, cfg, logger)
package health
