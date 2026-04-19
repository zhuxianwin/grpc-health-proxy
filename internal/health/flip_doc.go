// Package health – flip
//
// FlipChecker wraps any Checker and inverts its Status on demand.
// It is intended for testing and chaos-engineering scenarios where
// you want to temporarily force a healthy service to appear unhealthy
// (or vice-versa) without changing the underlying service.
//
// Usage:
//
//	base := health.NewChecker(addr, opts)
//	fc  := health.NewFlipChecker(base, health.DefaultFlipConfig())
//	fc.Toggle() // enable inversion
//	// ... fc.Toggle() again to restore normal behaviour
package health
