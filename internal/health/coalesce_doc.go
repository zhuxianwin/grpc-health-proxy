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
// # Configuration
//
// CoalesceConfig exposes a single Window field (time.Duration). Smaller values
// reduce latency at the cost of less coalescing; larger values increase
// coalescing but add tail latency for callers that just miss an in-flight
// request. The default window of 50 ms is a reasonable starting point for
// Kubernetes liveness/readiness probes, which typically fire every 10 s.
//
// # Concurrency
//
// The coalesce checker is safe for concurrent use. Each distinct service name
// is tracked independently, so checks for different services never block one
// another.
package health
