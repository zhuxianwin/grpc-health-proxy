// Package health — sentinel
//
// SentinelChecker absorbs cascading failures by tripping after a configurable
// number of consecutive unhealthy results for a given service. While tripped
// the checker returns a synthetic healthy status so that upstream systems are
// shielded from repeated error propagation. After ResetAfter elapses the
// sentinel resets and normal delegation resumes.
//
// Typical use:
//
//	checker := health.NewSentinelChecker(inner, health.DefaultSentinelConfig())
//
// The companion HTTP handler (NewSentinelStatusHandler) exposes per-service
// trip state as JSON for debugging and observability.
package health
