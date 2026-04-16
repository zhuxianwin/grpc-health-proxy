// Package health — quota
//
// QuotaChecker enforces a maximum number of health-check calls within a
// rolling time window, protecting upstream gRPC services from being
// overwhelmed by probe traffic during incident conditions.
//
// Usage:
//
//	cfg := health.DefaultQuotaConfig()   // 60 calls / minute
//	cfg.MaxCalls = 30
//	cfg.Window   = 30 * time.Second
//
//	var c health.Checker = health.NewChecker(addr, creds, logger)
//	c = health.NewQuotaChecker(c, cfg, logger)
//
// When the quota is exhausted the checker returns a non-nil Err and logs a
// warning at WARN level. The counter resets automatically after each Window.
//
// NewQuotaStatusHandler exposes the current counter and exhaustion state as
// a JSON HTTP endpoint suitable for debugging dashboards.
package health
