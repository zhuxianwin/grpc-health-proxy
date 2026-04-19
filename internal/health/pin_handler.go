package health

import (
	"encoding/json"
	"net/http"
)

type pinStatusResponse struct {
	Pinned bool   `json:"pinned"`
	Status string `json:"status,omitempty"`
}

// NewPinStatusHandler returns an HTTP handler that reports whether a
// PinChecker currently has a result pinned, and what that result is.
func NewPinStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, ok := c.(*PinChecker)
		if !ok {
			http.Error(w, "checker is not a PinChecker", http.StatusBadRequest)
			return
		}
		resp := pinStatusResponse{}
		if pin := pc.Pinned(); pin != nil {
			resp.Pinned = true
			resp.Status = pin.Status.String()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
