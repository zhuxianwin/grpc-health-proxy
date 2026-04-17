// Package health — batch checker.
//
// BatchChecker fans out health checks to multiple targets concurrently,
// honouring a configurable concurrency ceiling so that a large fleet of
// services does not overwhelm the host process or the network.
//
// Usage:
//
//	checkers := map[string]health.Checker{
//		"payments": health.NewChecker(conn1, log),
//		"orders":   health.NewChecker(conn2, log),
//	}
//	bc := health.NewBatchChecker(checkers, health.DefaultBatchConfig())
//	results := bc.CheckAll(ctx)
//
// The companion BatchStatusHandler exposes the aggregated results as JSON
// over HTTP, returning 200 when all targets are healthy and 503 otherwise.
package health
