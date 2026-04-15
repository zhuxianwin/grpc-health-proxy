// Package health provides gRPC health-check primitives used by the proxy.
//
// The package is structured around four main types:
//
//   - Checker performs a single gRPC health check against a remote service.
//   - Cache stores recent check results with a configurable TTL so that the
//     HTTP handler can answer without blocking on every request.
//   - Watcher runs a background polling loop that periodically calls Checker
//     and refreshes the Cache, decoupling the HTTP response latency from the
//     gRPC round-trip time.
//   - Handler is an http.Handler that reads from the Cache and writes an
//     appropriate HTTP status code (200 / 503) to the caller.
//
// Typical usage:
//
//	cache   := health.NewCache(cfg.CacheTTL)
//	checker := health.NewChecker(cfg.GRPCAddr, logger)
//	watcher := health.NewWatcher(checker, cache, cfg.Services, cfg.PollInterval, logger)
//	watcher.Start(ctx)
//	http.Handle("/healthz", health.NewHandler(cache, cfg.Services))
package health
