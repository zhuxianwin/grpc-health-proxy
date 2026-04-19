// Package health – cooldown checker
//
// CooldownChecker prevents a flapping service from being re-checked
// immediately after it returns unhealthy. Once a service is marked
// unhealthy the checker caches that result and returns it for the
// configured Duration without delegating to the inner Checker.
//
// After the cooldown period expires the next call is forwarded to the
// inner Checker as normal. If the service is still unhealthy a new
// cooldown window begins.
//
// Usage:
//
//	cc := health.NewCooldownChecker(
//		innerChecker,
//		health.CooldownConfig{Duration: 30 * time.Second},
//		logger,
//	)
//
// Zero-value CooldownConfig fields are replaced with DefaultCooldownConfig
// values automatically.
package health
