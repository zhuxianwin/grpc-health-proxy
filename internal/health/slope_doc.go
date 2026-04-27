// Package health — slope checker
//
// SlopeChecker tracks the healthy-rate trend over a sliding window of recent
// check results and returns Unhealthy when the least-squares slope of that
// rate falls below a configurable threshold.
//
// This is useful for catching gradual degradation that has not yet crossed a
// fixed healthy/unhealthy boundary but is clearly trending in the wrong
// direction.
//
// Example
//
//	cfg := health.DefaultSlopeConfig() // window=10, threshold=-0.3
//	checker := health.NewSlopeChecker(inner, cfg, logger)
package health
