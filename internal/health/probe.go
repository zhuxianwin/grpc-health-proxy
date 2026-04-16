package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// ProbeConfig controls behaviour of the HTTP probe handler.
type ProbeConfig struct {
	// Timeout is the maximum time allowed for a single health check.
	Timeout time.Duration
	// ServiceName is the gRPC service to probe. Empty string probes the
	// server-level health (empty service name in the gRPC health protocol).
	ServiceName string
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
func DefaultProbeConfig() ProbeConfig {
	return ProbeConfig{
		Timeout:     5 * time.Second,
		ServiceName: "",
	}
}

// NewProbeHandler returns an http.Handler that performs a single health check
// against the provided Checker and writes an appropriate HTTP status code.
//
// 200 OK      – service is healthy
// 503 Service Unavailable – service is unhealthy or the check errored
func NewProbeHandler(checker Checker, cfg ProbeConfig, logger *slog.Logger) http.Handler {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultProbeConfig().Timeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
		defer cancel()

		result := checker.Check(ctx, cfg.ServiceName)

		if result.Err != nil {
			logger.WarnContext(ctx, "probe check error",
				"service", cfg.ServiceName,
				"err", result.Err,
			)
			http.Error(w, "health check error: "+result.Err.Error(), http.StatusServiceUnavailable)
			return
		}

		if result.Status != StatusHealthy {
			logger.InfoContext(ctx, "probe unhealthy",
				"service", cfg.ServiceName,
				"status", result.Status.String(),
			)
			http.Error(w, "service unhealthy: "+result.Status.String(), http.StatusServiceUnavailable)\n		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
