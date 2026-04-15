package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type fallbackStatusResponse struct {
	PrimaryOK    bool   `json:"primary_ok"`
	FallbackUsed bool   `json:"fallback_used"`
	Service      string `json:"service"`
	Source       string `json:"source"`
}

// NewFallbackStatusHandler returns an http.Handler that reports whether the
// last check for a service was served by the fallback checker.
// It accepts a FallbackChecker and the target service name.
func NewFallbackStatusHandler(fc *FallbackChecker, service string, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := fc.Check(r.Context(), service)

		fallbackUsed := res.Source == "fallback"
		primaryOK := !fallbackUsed && res.Err == nil

		resp := fallbackStatusResponse{
			PrimaryOK:    primaryOK,
			FallbackUsed: fallbackUsed,
			Service:      service,
			Source:       res.Source,
		}

		statusCode := http.StatusOK
		if res.Status != StatusHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode fallback status response", "error", err)
		}
	})
}
