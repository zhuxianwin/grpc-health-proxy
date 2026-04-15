package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RateLimitStatusHandler exposes the current rate-limit configuration and
// a per-service token availability indicator over HTTP.
type RateLimitStatusHandler struct {
	checkers map[string]*RateLimitedChecker
	logger   *slog.Logger
}

// NewRateLimitStatusHandler returns an HTTP handler that reports rate-limit
// status for all registered RateLimitedCheckers.
func NewRateLimitStatusHandler(checkers map[string]*RateLimitedChecker, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RateLimitStatusHandler{checkers: checkers, logger: logger}
}

type rateLimitServiceStatus struct {
	Service        string  `json:"service"`
	TokenAvailable bool    `json:"token_available"`
	Rate           float64 `json:"rate_per_second"`
	Burst          float64 `json:"burst"`
}

func (h *RateLimitStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statuses := make([]rateLimitServiceStatus, 0, len(h.checkers))
	for name, rl := range h.checkers {
		// Probe without consuming a token by inspecting bucket state.
		avail := rl.bucket.tokens >= 1
		statuses = append(statuses, rateLimitServiceStatus{
			Service:        name,
			TokenAvailable: avail,
			Rate:           rl.bucket.rate * float64(1e9), // convert back to rps
			Burst:          rl.bucket.max,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		h.logger.Error("rate-limit handler: encode error", "err", err)
	}
}
