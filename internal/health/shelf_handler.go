package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type shelfHandlerPayload struct {
	Service   string `json:"service"`
	Shelved   bool   `json:"shelved"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

// NewShelfStatusHandler returns an http.Handler that reports which services
// currently have a shelved (cached healthy) result and when each expires.
// It responds 200 with a JSON array regardless of shelf contents.
func NewShelfStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*shelfChecker)
		if !ok {
			http.Error(w, "checker is not a ShelfChecker", http.StatusBadRequest)
			return
		}

		now := time.Now()
		sc.mu.Lock()
		payload := make([]shelfHandlerPayload, 0, len(sc.shelf))
		for svc, entry := range sc.shelf {
			if now.Before(entry.expiresAt) {
				payload = append(payload, shelfHandlerPayload{
					Service:   svc,
					Shelved:   true,
					ExpiresIn: entry.expiresAt.Sub(now).Round(time.Millisecond).String(),
				})
			}
		}
		sc.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}
