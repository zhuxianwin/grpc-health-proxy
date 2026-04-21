// Package health — debounce
//
// NewDebounceChecker wraps a Checker so that repeated calls for the same
// service within a configurable Window return the previously observed Result
// without hitting the upstream gRPC endpoint again.
//
// This is useful when many Kubernetes probes fire simultaneously: only the
// first call in each window pays the cost of a real gRPC round-trip; all
// subsequent callers within that window receive the cached answer instantly.
//
// Example:
//
//	checker := health.NewDebounceChecker(
//		baseChecker,
//		health.DefaultDebounceConfig(),
//		logger,
//	)
package health
