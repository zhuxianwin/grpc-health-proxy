// Package health — fallback checker
//
// FallbackChecker provides resilience by delegating to a secondary Checker
// when the primary Checker returns an error (e.g. the upstream gRPC service
// is temporarily unreachable).
//
// Behaviour summary:
//
//   - Primary returns healthy  → result returned as-is.
//   - Primary returns unhealthy (no error) → result returned as-is; the
//     fallback is NOT consulted so that a deliberate SERVING=NOT_SERVING
//     response is honoured.
//   - Primary returns an error → fallback is invoked; the returned Result
//     has Source set to "fallback" for observability.
//
// Usage:
//
//	primary  := health.NewChecker(addr, "my-service", creds, logger)
//	fallback := health.NewChecker(addrBackup, "my-service", creds, logger)
//	fc := health.NewFallbackChecker(primary, fallback, logger)
package health
