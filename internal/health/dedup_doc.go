// Package health — dedup
//
// DedupChecker collapses concurrent in-flight health-check requests for the
// same service name into a single upstream gRPC call using a singleflight
// group. This is particularly valuable when many Kubernetes probes fire at
// the same instant, preventing a cascade of redundant RPCs to the target
// service.
//
// Usage:
//
//	base := health.NewChecker(target, creds, log)
//	checker := health.NewDedupChecker(base)
//	// checker can now be used wherever a Checker is expected.
//
// DedupChecker is safe for concurrent use. The Forget method removes a
// service's in-flight record, which is useful in tests or after a known
// topology change where a fresh probe is desired immediately.
package health
