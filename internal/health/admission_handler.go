package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type admissionStatusResponse struct {
	Active  int64 `json:"active"`
	Max     int64 `json:"max"`
	Available int64 `json:"available"`
}

// NewAdmissionStatusHandler returns an http.Handler that reports the current
// concurrency state of the given admissionChecker. If checker is not an
// *admissionChecker the handler responds with 501 Not Implemented.
func NewAdmissionStatusHandler(checker Checker, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac, ok := checker.(*admissionChecker)
		if !ok {
			http.Error(w, "checker is not an admission checker", http.StatusNotImplemented)
			return
		}

		active := ac.active.Load()
		resp := admissionStatusResponse{
			Active:    active,
			Max:       ac.max,
			Available: ac.max - active,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode admission status", "err", err)
		}
	})
}
