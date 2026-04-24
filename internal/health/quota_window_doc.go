// Package health — quota_window
//
// QuotaWindowChecker enforces a sliding-window call-rate quota on health
// checks. Unlike the fixed-window QuotaChecker, this implementation tracks
// individual call timestamps and evicts those that fall outside the rolling
// window, giving a smoother enforcement boundary.
//
// Usage:
//
//	cfg := health.DefaultQuotaWindowConfig()
//	cfg.Window   = 30 * time.Second
//	cfg.MaxCalls = 10
//	checker := health.NewQuotaWindowChecker(inner, cfg)
//
// When the per-service quota is exhausted the checker returns
// StatusUnknown with ErrQuotaExceeded without forwarding the call to the
// wrapped checker.
package health
