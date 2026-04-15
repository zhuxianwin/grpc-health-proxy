package health_test

import (
	"net/http"
	"testing"

	"github.com/yourorg/grpc-health-proxy/internal/health"
)

func makeResults(statuses ...health.Status) map[string]health.Result {
	m := make(map[string]health.Result, len(statuses))
	for i, s := range statuses {
		key := string(rune('a' + i))
		m[key] = health.Result{Status: s}
	}
	return m
}

func TestAggregateStatus_AllHealthy(t *testing.T) {
	results := makeResults(health.StatusHealthy, health.StatusHealthy)
	if got := health.AggregateStatus(results); got != http.StatusOK {
		t.Errorf("expected 200, got %d", got)
	}
}

func TestAggregateStatus_OneUnhealthy(t *testing.T) {
	results := makeResults(health.StatusHealthy, health.StatusUnhealthy)
	if got := health.AggregateStatus(results); got != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", got)
	}
}

func TestAggregateStatus_Empty(t *testing.T) {
	if got := health.AggregateStatus(nil); got != http.StatusOK {
		t.Errorf("expected 200 for empty results, got %d", got)
	}
}

func TestBuildAggregateResponse_Overall(t *testing.T) {
	results := makeResults(health.StatusHealthy, health.StatusUnhealthy)
	resp := health.BuildAggregateResponse(results)
	if resp.Overall != health.StatusUnhealthy.String() {
		t.Errorf("expected overall unhealthy, got %s", resp.Overall)
	}
	if len(resp.Targets) != 2 {
		t.Errorf("expected 2 targets in response, got %d", len(resp.Targets))
	}
}

func TestBuildAggregateResponse_AllHealthy(t *testing.T) {
	results := makeResults(health.StatusHealthy)
	resp := health.BuildAggregateResponse(results)
	if resp.Overall != health.StatusHealthy.String() {
		t.Errorf("expected overall healthy, got %s", resp.Overall)
	}
}
