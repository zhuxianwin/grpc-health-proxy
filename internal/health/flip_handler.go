package health

import (
	"encoding/json"
	"net/http"
)

type flipStatusHandler struct {
	f *flipChecker
}

// NewFlipStatusHandler returns an http.Handler that reports the current
// flip state and exposes a POST /toggle endpoint via query param action=toggle.
func NewFlipStatusHandler(f *flipChecker) http.Handler {
	return &flipStatusHandler{f: f}
}

func (h *flipStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Query().Get("action") == "toggle" {
		newState := h.f.Toggle()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"active": newState})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"active": h.f.Active()})
}
