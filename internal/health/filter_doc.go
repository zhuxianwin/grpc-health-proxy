// Package health provides gRPC health check primitives.
//
// # Filter
//
// FilterChecker restricts health checks to an explicit allow list of service
// names. Services not present in the list are immediately returned as unhealthy
// without forwarding the call to the inner Checker.
//
// When AllowedServices is empty, all services are forwarded (allow-all mode).
//
// Example:
//
//	cfg := health.DefaultFilterConfig()
//	cfg.AllowedServices = map[string]struct{}{
//		"my.Service": {},
//	}
//	checker := health.NewFilterChecker(inner, cfg)
package health
