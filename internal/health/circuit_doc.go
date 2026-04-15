// Package health provides gRPC health checking primitives used by the
// grpc-health-proxy sidecar.
//
// # Circuit Breaker
//
// NewCircuitBreaker wraps any Checker with a three-state circuit breaker:
//
//   - Closed  – normal operation; every call is forwarded to the inner Checker.
//   - Open    – the upstream is considered down; calls fail immediately without
//     contacting the upstream, reducing load during outages.
//   - HalfOpen – after OpenDuration has elapsed the circuit allows a probe
//     request through. Consecutive successes (>= SuccessThreshold) close the
//     circuit again; any failure re-opens it.
//
// CircuitConfig controls all thresholds. Use DefaultCircuitConfig for
// production-ready defaults.
//
// The companion NewCircuitStatusHandler exposes the current circuit state as a
// JSON HTTP endpoint, suitable for mounting alongside the standard health and
// metrics routes.
package health
