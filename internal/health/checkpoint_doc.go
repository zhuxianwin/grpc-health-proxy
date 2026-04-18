// Package health provides gRPC health check primitives.
//
// # Checkpoint
//
// CheckpointStore records the most recent health result for each service so
// that operators can inspect historical state without querying the upstream
// gRPC target directly.
//
// NewCheckpointChecker wraps any Checker and writes to the store at most once
// per MinInterval to avoid lock contention under high probe rates.
//
// NewCheckpointStatusHandler exposes the store contents as a JSON HTTP
// endpoint suitable for mounting on a debug or admin mux.
package health
