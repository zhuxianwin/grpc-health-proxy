package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type hedgeStatusHandler struct {
	cfg    HedgeConfig
	logger *slog.Logger
}

// NewHedgeStatusHandler returns an HTTP handler that reports the current hedge
// configuration as JSON. Useful for debugging the running sidecar.
func NewHedgeStatusHandler(cfg HedgeConfig, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &hedgeStatusHandler{cfg: cfg, logger: logger}
}

func (h *hedgeStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		DelayMs   int64 `json:"delay_ms"`
		MaxHedged int   `json:"max_hedged"`
	}{
		DelayMs:   h.cfg.Delay.Milliseconds(),
		MaxHedged: h.cfg.MaxHedged,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("hedge status: failed to encode response", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
