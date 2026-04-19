package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// NewSuppressStatusHandler returns an HTTP handler that reports the current
// suppression set of a SuppressChecker as JSON.
func NewSuppressStatusHandler(c Checker, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*SuppressChecker)
		if !ok {
			http.Error(w, "not a SuppressChecker", http.StatusBadRequest)
			return
		}
		sc.mu.RLock()
		statuses := make([]string, 0, len(sc.set))
		for st := range sc.set {
			statuses = append(statuses, st.String())
		}
		sc.mu.RUnlock()

		payload := map[string]any{
			"suppressed_statuses": statuses,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Error("suppress handler encode", "err", err)
		}
	})
}
