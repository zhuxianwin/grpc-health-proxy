// Package health — rate-limit module
//
// RateLimitedChecker wraps any Checker with a token-bucket rate limiter so
// that downstream gRPC services are not overwhelmed by rapid probe bursts.
//
// # Usage
//
//	cfg := health.DefaultRateLimitConfig()
//	cfg.MaxChecksPerSecond = 5
//	cfg.Burst = 2
//
//	base := health.NewChecker(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	rl   := health.NewRateLimitedChecker(base, "my-service", cfg)
//
//	// rl implements health.Checker and can be composed with retry/timeout/circuit.
//
// # HTTP status endpoint
//
// NewRateLimitStatusHandler returns an http.Handler that reports the current
// token-availability state for each registered RateLimitedChecker.  Mount it
// at a debug path such as /debug/ratelimit.
package health
