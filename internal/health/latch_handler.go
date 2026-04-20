package health

import (
	"encoding/json"
	"net/http"
)

type latchStatusResponse struct {
	Tripped   bool `json:"tripped"`
	Threshold int  `json:"threshold"`
}

// latchStatusChecker is the interface satisfied by latchChecker for the handler.
type latchStatusChecker interface {
	Tripped() bool
	Reset()
}

// NewLatchStatusHandler returns an HTTP handler that reports latch state
// and accepts POST requests to manually reset it.
//
//	GET  /healthz/latch  → JSON status
//	POST /healthz/latch  → resets the latch, returns updated status
func NewLatchStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lc, ok := c.(latchStatusChecker)
		if !ok {
			http.Error(w, "checker is not a latch", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodPost {
			lc.Reset()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(latchStatusResponse{
			Tripped: lc.Tripped(),
		})
	})
}
