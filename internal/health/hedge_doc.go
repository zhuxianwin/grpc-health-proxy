// Package health provides gRPC health check primitives.
//
// # Hedge Checker
//
// NewHedgeChecker wraps an inner Checker and fires a speculative
// ("hedged") duplicate request after a configurable delay. Whichever
// response arrives first is returned to the caller; the slower
// in-flight call is cancelled via context.
//
// This pattern reduces tail latency at the cost of occasional
// duplicate backend load. It is most useful when a small fraction of
// health checks are slow due to transient network hiccups.
//
// Configuration:
//
//	Delay      – how long to wait before firing the hedge request.
//	MaxHedges  – maximum concurrent hedge requests (default 1).
//
Usage:
//
//	cfg := health.DefaultHedgeConfig()
//	cfg.Delay = 50 * time.Millisecond
//	checker := health.NewHedgeChecker(inner, cfg, logger)
package health
