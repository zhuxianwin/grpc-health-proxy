// Package health – mirror checker
//
// MirrorChecker duplicates every health check to one or more mirror targets
// while always returning the primary checker's result to the caller.
//
// Use cases:
//   - Shadow-testing a new gRPC backend without affecting probe outcomes.
//   - Warming caches on standby replicas in parallel with the live target.
//
// Mirror checks run concurrently in background goroutines. Any error from a
// mirror is logged but never propagates to the caller.
package health
