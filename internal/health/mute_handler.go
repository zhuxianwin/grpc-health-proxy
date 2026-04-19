package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type muteStatusResponse struct {
	Muted      bool      `json:"muted"`
	MutedUntil time.Time `json:"muted_until,omitempty"`
}

// NewMuteStatusHandler returns an HTTP handler that reports mute state
// and accepts POST /?action=mute|unmute to control it.
func NewMuteStatusHandler(mc *MuteChecker, log *slog.Logger) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			action := r.URL.Query().Get("action")
			switch action {
			case "mute":
				mc.Mute()
			case "unmute":
				mc.Unmute()
			default:
				http.Error(w, "unknown action", http.StatusBadRequest)
				return
			}
		}
		mc.mu.Lock()
		until := mc.mutedUntil
		mc.mu.Unlock()
		resp := muteStatusResponse{
			Muted:      time.Now().Before(until),
			MutedUntil: until,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("failed to encode mute status", "err", err)
		}
	})
}
