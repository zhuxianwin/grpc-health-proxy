// Package metrics centralises all Prometheus instrumentation for the
// grpc-health-proxy sidecar.
//
// # Exported metrics
//
//   - grpc_health_proxy_health_checks_total — counter, labelled by
//     "service" and "status" (healthy | unhealthy | unreachable).
//
//   - grpc_health_proxy_health_check_duration_seconds — histogram,
//     labelled by "service", tracking round-trip latency of each
//     upstream gRPC health check call.
//
//   - grpc_health_proxy_http_requests_total — counter, labelled by
//     "path" and "code", tracking inbound HTTP requests to the proxy
//     itself (e.g. /healthz, /readyz, /livez).
//
// # Usage
//
// Import this package from any layer that needs to record an
// observation.  The metrics are registered with the default
// Prometheus registry via promauto on package initialisation.
//
// Expose the /metrics endpoint by mounting [Handler] on your HTTP mux:
//
//	mux.Handle("/metrics", metrics.Handler())
package metrics
