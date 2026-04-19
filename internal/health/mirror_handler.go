package health

import (
	"encoding/json"
	"net/http"
)

type mirrorStatusResponse struct {
	MirrorCount int `json:"mirror_count"`
}

// NewMirrorStatusHandler returns an HTTP handler that reports the number of
// configured mirror targets for a mirrorChecker.
func NewMirrorStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc, ok := c.(*mirrorChecker)
		if !ok {
			http.Error(w, "not a mirror checker", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mirrorStatusResponse{
			MirrorCount: len(mc.mirrors),
		})
	})
}
