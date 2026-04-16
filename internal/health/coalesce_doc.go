// Package health provides gRPC health-check primitives used by the proxy.
//
// # Coalesce
//
// NewCoalesceChecker wraps any Checker and collapses concurrent calls for the
// same service name into a single upstream request. Callers that arrive within
// the configured Window block until the in-flight check completes and then
// receive its result, reducing load on the upstream gRPC service during bursts
// of readiness-probe traffic.
//
// Usage:
//
//	cfg := health.DefaultCoalesceConfig()  // 50 ms window
//	checker := health.NewCoalesceChecker(base, cfg, logger)
//
// The coalesce checker is safe for concurrent use.
package health
