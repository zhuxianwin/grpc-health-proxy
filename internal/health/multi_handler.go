package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// MultiHandler is an HTTP handler that checks multiple gRPC targets and
// returns an aggregated JSON response.
type MultiHandler struct {
	mc      *MultiChecker
	service string
	timeout time.Duration
	log     *slog.Logger
}

// NewMultiHandler constructs a MultiHandler.
func NewMultiHandler(mc *MultiChecker, service string, timeout time.Duration, log *slog.Logger) *MultiHandler {
	return &MultiHandler{mc: mc, service: service, timeout: timeout, log: log}
}

func (h *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	results := h.mc.CheckAll(ctx, h.service)
	resp := BuildAggregateResponse(results)
	status := AggregateStatus(results)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("failed to encode aggregate response", "err", err)
	}
}
