package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// BatchStatusHandler serves a JSON summary of all batch check results.
type BatchStatusHandler struct {
	checker *BatchChecker
	log     *slog.Logger
}

// NewBatchStatusHandler creates an HTTP handler backed by a BatchChecker.
func NewBatchStatusHandler(bc *BatchChecker, log *slog.Logger) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return &BatchStatusHandler{checker: bc, log: log}
}

func (h *BatchStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	results := h.checker.CheckAll(r.Context())

	type entry struct {
		Target  string `json:"target"`
		Status  string `json:"status"`
		Healthy bool   `json:"healthy"`
		Error   string `json:"error,omitempty"`
	}

	overall := StatusHealthy
	entries := make([]entry, 0, len(results))
	for _, br := range results {
		e := entry{
			Target:  br.Target,
			Status:  br.Result.Status.String(),
			Healthy: br.Result.Status == StatusHealthy,
		}
		if br.Result.Err != nil {
			e.Error = br.Result.Err.Error()
		}
		if br.Result.Status != StatusHealthy {
			overall = StatusUnhealthy
		}
		entries = append(entries, e)
	}

	body := map[string]any{
		"overall": overall.String(),
		"checks":  entries,
	}

	code := httpStatusCode(overall)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.log.Error("batch handler encode", "err", err)
	}
}

// httpStatusCode maps a health Status to an HTTP status code.
// Healthy targets return 200 OK; anything else returns 503 Service Unavailable.
func httpStatusCode(s Status) int {
	if s == StatusHealthy {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
}
