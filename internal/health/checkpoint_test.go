package health

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckpointStore_RecordAndLatest(t *testing.T) {
	s := NewCheckpointStore()
	_, ok := s.Latest("svc")
	if ok {
		t.Fatal("expected no entry")
	}
	s.Record("svc", Result{Status: StatusHealthy})
	e, ok := s.Latest("svc")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Status != StatusHealthy {
		t.Fatalf("got %v", e.Status)
	}
}

func TestCheckpointStore_All(t *testing.T) {
	s := NewCheckpointStore()
	s.Record("a", Result{Status: StatusHealthy})
	s.Record("b", Result{Status: StatusUnhealthy})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
}

func TestCheckpointChecker_RecordsResult(t *testing.T) {
	store := NewCheckpointStore()
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	cfg := CheckpointConfig{MinInterval: 0}
	checker := NewCheckpointChecker(inner, store, cfg, nil)
	checker.Check(context.Background(), "svc")
	_, ok := store.Latest("svc")
	if !ok {
		t.Fatal("expected checkpoint to be recorded")
	}
}

func TestCheckpointChecker_MinIntervalSuppresses(t *testing.T) {
	store := NewCheckpointStore()
	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusHealthy}
	})
	cfg := CheckpointConfig{MinInterval: time.Hour}
	checker := NewCheckpointChecker(inner, store, cfg, nil)
	checker.Check(context.Background(), "svc")
	checker.Check(context.Background(), "svc")
	// Both calls hit inner but only first should checkpoint
	if calls != 2 {
		t.Fatalf("expected 2 inner calls, got %d", calls)
	}
	// Store should have exactly one entry
	if len(store.All()) != 1 {
		t.Fatalf("expected 1 checkpoint")
	}
}

func TestCheckpointHandler_ReturnsJSON(t *testing.T) {
	store := NewCheckpointStore()
	store.Record("svc", Result{Status: StatusHealthy})
	h := NewCheckpointStatusHandler(store)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["count"].(float64) != 1 {
		t.Fatalf("expected count 1")
	}
}

func TestCheckpointHandler_ContentType(t *testing.T) {
	h := NewCheckpointStatusHandler(NewCheckpointStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}
