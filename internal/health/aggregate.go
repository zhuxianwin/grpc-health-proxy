package health

import "net/http"

// AggregateStatus derives an overall HTTP status code from a map of results.
// Returns 200 if all targets are healthy, 503 otherwise.
func AggregateStatus(results map[string]Result) int {
	for _, r := range results {
		if r.Status != StatusHealthy {
			return http.StatusServiceUnavailable
		}
	}
	return http.StatusOK
}

// AggregateResponse is a JSON-serialisable summary of a multi-target check.
type AggregateResponse struct {
	Overall string            `json:"overall"`
	Targets map[string]string `json:"targets"`
}

// BuildAggregateResponse converts raw results into an AggregateResponse.
func BuildAggregateResponse(results map[string]Result) AggregateResponse {
	targets := make(map[string]string, len(results))
	for t, r := range results {
		targets[t] = r.Status.String()
	}
	overall := StatusHealthy.String()
	for _, r := range results {
		if r.Status != StatusHealthy {
			overall = StatusUnhealthy.String()
			break
		}
	}
	return AggregateResponse{Overall: overall, Targets: targets}
}
