// Package health – burst checker
//
// # Burst
//
// NewBurstChecker wraps a Checker and enforces a maximum call rate within a
// rolling time window.  When the number of Check invocations exceeds
// BurstConfig.MaxBurst within BurstConfig.Window the checker returns
// StatusUnhealthy without delegating to the inner checker.
//
// This is useful for protecting downstream services from health-check storms
// during rapid pod restarts or aggressive Kubernetes probe configurations.
//
// # Status handler
//
// NewBurstStatusHandler exposes the current configuration and live call count
// as a JSON HTTP endpoint so operators can observe burst pressure in real time.
package health
