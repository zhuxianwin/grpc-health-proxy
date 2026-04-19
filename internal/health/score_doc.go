// Package health provides gRPC health check utilities.
//
// # Score Checker
//
// The score checker aggregates multiple downstream checkers into a single
// weighted health score. Each entry has a service name, a numeric weight,
// and an inner Checker. The overall score is:
//
//	score = sum(weight_i for healthy_i) / sum(weight_i)
//
// The result is healthy when score >= Threshold (default 0.5).
//
// Example usage:
//
//	entries := []health.scoreEntry{
//	    health.NewScoreEntry("db",    3.0, dbChecker),
//	    health.NewScoreEntry("cache", 1.0, cacheChecker),
//	}
//	cfg := health.DefaultScoreConfig()
//	cfg.Threshold = 0.75
//	checker := health.NewScoreChecker(cfg, entries, logger)
package health
