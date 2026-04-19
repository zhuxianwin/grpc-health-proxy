// Package health — sticky checker
//
// NewStickyChecker wraps any Checker so that once a service is observed as
// unhealthy, that unhealthy result is "stuck" for a configurable TTL.  This
// prevents flapping: a brief recovery does not immediately mark the service
// healthy again.
//
// Once the TTL expires the inner checker is consulted on the next call.  A
// successful healthy result clears the sticky entry immediately.
//
// Use NewStickyStatusHandler to expose current sticky state over HTTP.
package health
