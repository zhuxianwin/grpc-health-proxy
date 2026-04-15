package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// CircuitStatusResponse is the JSON body returned by the circuit status endpoint.
type CircuitStatusResponse struct {
	Service string `json:"service"`
	State   string `json:"circuit_state"`
}

func circuitStateName(s CircuitState) string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// NewCircuitStatusHandler returns an HTTP handler that exposes the circuit
// breaker state for a named service. The circuitBreaker must be cast from
// the Checker returned by NewCircuitBreaker.
func NewCircuitStatusHandler(service string, cb Checker, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		breaker, ok := cb.(*circuitBreaker)
		if !ok {
			http.Error(w, "not a circuit breaker", http.StatusInternalServerError)
			return
		}
		breaker.mu.Lock()
		state := breaker.resolveState()
		breaker.mu.Unlock()

		resp := CircuitStatusResponse{
			Service: service,
			State:   circuitStateName(state),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("circuit status encode", "err", err)
		}
	})
}
