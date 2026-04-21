package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type debounceStatusResponse struct {
	Window string `json:"window"`
}

// NewDebounceStatusHandler returns an HTTP handler that reports the configured
// debounce window for the given checker. If c is not a *debounceChecker the
// handler responds with 400.
func NewDebounceStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, ok := c.(*debounceChecker)
		if !ok {
			http.Error(w, "not a debounce checker", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(debounceStatusResponse{
			Window: db.cfg.Window.Round(time.Millisecond).String(),
		})
	})
}
