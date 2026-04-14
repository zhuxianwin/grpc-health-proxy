package health

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Handler is an HTTP handler that exposes a gRPC health check endpoint.
type Handler struct {
	checker *Checker
	service string
}

// NewHandler creates an HTTP handler wrapping the given Checker.
func NewHandler(addr string, service string, timeout time.Duration) *Handler {
	return &Handler{
		checker: NewChecker(addr, timeout),
		service: service,
	}
}

type response struct {
	Status  string `json:"status"`
	Service string `json:"service,omitempty"`
}

// ServeHTTP handles HTTP readiness probe requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := h.checker.Check(context.Background(), h.service)
	if err != nil {
		log.Printf("health check error: %v", err)
	}

	body := response{
		Status:  status.String(),
		Service: h.service,
	}

	w.Header().Set("Content-Type", "application/json")

	if status == StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
