// Package health — budget
//
// BudgetChecker enforces an error budget over a sliding time window.
//
// It tracks the fraction of failed health checks for each service. Once that
// fraction exceeds the configured ErrorBudget threshold (and the window
// contains at least MinSamples entries) the checker returns Unhealthy with
// ErrBudgetExhausted until the error rate drops back below the threshold.
//
// Example:
//
//	cfg := health.DefaultBudgetConfig()
//	cfg.ErrorBudget = 0.05 // 5 % error budget
//	checker := health.NewBudgetChecker(inner, cfg)
package health
