package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchHandler_AllHealthy(t *testing.T) {
	checkers := map[string]Checker{
		"svc-a": &fixedChecker{result: Result{Status: StatusHealthy}},
	}
	bc := NewBatchChecker(checkers, DefaultBatchConfig())
	h := NewBatchStatusHandler(bc, nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["overall"] != "healthy" {
		t.Errorf("expected overall healthy, got %v", body["overall"])
	}
}

func TestBatchHandler_OneUnhealthy(t *testing.T) {
	checkers := map[string]Checker{
		"svc-a": &fixedChecker{result: Result{Status: StatusHealthy}},
		"svc-b": &fixedChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("fail")}},
	}
	bc := NewBatchChecker(checkers, DefaultBatchConfig())
	h := NewBatchStatusHandler(bc, nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["overall"] != "unhealthy" {
		t.Errorf("expected overall unhealthy, got %v", body["overall"])
	}
}

func TestBatchHandler_ContentType(t *testing.T) {
	bc := NewBatchChecker(map[string]Checker{}, DefaultBatchConfig())
	h := NewBatchStatusHandler(bc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected content-type: %s", ct)
	}
}
