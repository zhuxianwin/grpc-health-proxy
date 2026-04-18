package health

import (
	"encoding/json"
	"net/http"
)

type tagStatusHandler struct {
	checker *tagChecker
}

// NewTagStatusHandler returns an HTTP handler that reports the static tags
// configured on a tagChecker. Returns 400 if checker is not a *tagChecker.
func NewTagStatusHandler(c Checker) http.Handler {
	tc, ok := c.(*tagChecker)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not a tag checker", http.StatusBadRequest)
		})
	}
	return &tagStatusHandler{checker: tc}
}

func (h *tagStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags": h.checker.Tags(),
	})
}
