package health

import "time"

// Status represents the outcome of a single health check evaluation.
type Status int

const (
	// StatusHealthy indicates the upstream gRPC service reported SERVING.
	StatusHealthy Status = iota
	// StatusUnhealthy indicates the upstream reported NOT_SERVING or an
	// unexpected status value.
	StatusUnhealthy
	// StatusUnknown indicates the check could not be completed (e.g. the
	// upstream was unreachable or returned an error).
	StatusUnknown
)

// String returns a human-readable label for the status.
func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a health check together with metadata that is
// useful for logging and caching decisions.
type Result struct {
	// Service is the gRPC service name that was checked (empty string means
	// the server-level check).
	Service string
	// Status is the evaluated health status.
	Status Status
	// Err is non-nil when the check could not be completed successfully.
	Err error
	// CheckedAt records when the check was performed.
	CheckedAt time.Time
}

// Healthy returns true when the result represents a serving upstream.
func (r Result) Healthy() bool { return r.Status == StatusHealthy }
