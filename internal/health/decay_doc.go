// Package health — decay checker
//
// DecayChecker wraps any Checker and introduces exponential back-off
// cool-downs after a configurable number of consecutive failures.  While
// the service is in a decayed state calls are short-circuited and the
// last cached failure is returned immediately, avoiding a flood of
// outbound gRPC probes against an already-struggling upstream.
//
// Usage:
//
//	cfg := health.DefaultDecayConfig()
//	cfg.DecayAfter = 5
//	cfg.BaseDelay  = time.Second
//	checked := health.NewDecayChecker(inner, cfg, logger)
//
// The cool-down doubles with every additional failure (capped at
// MaxDelay) and resets to zero as soon as a successful result is
// observed.
package health
