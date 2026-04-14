// Package health provides gRPC health checking functionality for the
// grpc-health-proxy sidecar.
//
// It implements the gRPC Health Checking Protocol (grpc.health.v1) and exposes
// results over HTTP for use with Kubernetes readiness and liveness probes.
//
// The package is composed of three main components:
//
//   - Checker: dials a gRPC backend and calls the Health/Check RPC, returning
//     a boolean healthy status and any transport-level error.
//
//   - Cache: a short-lived in-memory result store that prevents every incoming
//     HTTP probe from fanning out to the upstream gRPC service. Entries expire
//     after a configurable TTL and can be invalidated explicitly.
//
//   - Handler: an http.Handler that wires the Checker and Cache together,
//     records Prometheus metrics, and writes an appropriate HTTP status code
//     (200 OK or 503 Service Unavailable) back to the caller.
package health
