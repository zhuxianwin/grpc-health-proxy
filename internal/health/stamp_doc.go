// Package health — stamp
//
// StampChecker wraps any Checker and attaches a "checked_at" timestamp to
// every Result's Metadata map. The timestamp is formatted as RFC3339 UTC.
//
// Usage:
//
//	cfg := health.DefaultStampConfig()
//	stamped := health.NewStampChecker(inner, cfg, logger)
//	res := stamped.Check(ctx, "my-service")
//	fmt.Println(res.Metadata["checked_at"]) // e.g. "2024-01-15T10:00:00Z"
//
// This is useful for debugging and audit logging, providing a clear record of
// when each health check was last evaluated.
package health
