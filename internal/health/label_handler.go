package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type labelStatusResponse struct {
	Labels map[string]string `json:"labels"`
}

// NewLabelStatusHandler returns an http.Handler that exposes the labels
// configured on a labelChecker as a JSON response.
func NewLabelStatusHandler(c Checker, log *slog.Logger) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lc, ok := c.(*labelChecker)
		if !ok {
			http.Error(w, "not a label checker", http.StatusBadRequest)
			return
		}
		resp := labelStatusResponse{Labels: lc.Labels()}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("label status handler encode error", "err", err)
		}
	})
}
