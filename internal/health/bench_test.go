package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fixedLatencyChecker struct {
	delay  time.Duration
	result Result
}

func (f *fixedLatencyChecker) Check(_ context.Context, _ string) Result {
	time.Sleep(f.delay)
	return f.result
}

func TestBench_RecordsLatency(t *testing.T) {
	inner := &fixedLatencyChecker{delay: 10 * time.Millisecond, result: Result{Status: StatusHealthy}}
	bc := NewBenchChecker(inner, DefaultBenchConfig())

	bc.Check(context.Background(), "svc")
	bc.Check(context.Background(), "svc")

	mean, max := bc.Stats()
	if mean <= 0 {
		t.Fatalf("expected positive mean, got %v", mean)
	}
	if max < mean {
		t.Fatalf("max %v should be >= mean %v", max, mean)
	}
}

func TestBench_WindowEviction(t *testing.T) {
	inner := &fixedLatencyChecker{result: Result{Status: StatusHealthy}}
	bc := NewBenchChecker(inner, BenchConfig{WindowSize: 3})

	for i := 0; i < 10; i++ {
		bc.Check(context.Background(), "svc")
	}

	bc.mu.Lock()
	n := len(bc.window)
	bc.mu.Unlock()

	if n != 3 {
		t.Fatalf("expected window size 3, got %d", n)
	}
}

func TestBench_EmptyStats(t *testing.T) {
	bc := NewBenchChecker(&fixedLatencyChecker{}, DefaultBenchConfig())
	mean, max := bc.Stats()
	if mean != 0 || max != 0 {
		t.Fatalf("expected zero stats on empty window, got mean=%v max=%v", mean, max)
	}
}

func TestBench_DefaultConfigOnZero(t *testing.T) {
	bc := NewBenchChecker(&fixedLatencyChecker{}, BenchConfig{})
	if bc.cfg.WindowSize != DefaultBenchConfig().WindowSize {
		t.Fatalf("expected default window size %d, got %d", DefaultBenchConfig().WindowSize, bc.cfg.WindowSize)
	}
}

func TestBenchHandler_ReturnsJSON(t *testing.T) {
	inner := &fixedLatencyChecker{delay: 5 * time.Millisecond, result: Result{Status: StatusHealthy}}
	bc := NewBenchChecker(inner, DefaultBenchConfig())
	bc.Check(context.Background(), "svc")

	h := NewBenchStatusHandler(bc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bench", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s benchStats
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Window != DefaultBenchConfig().WindowSize {
		t.Fatalf("unexpected window size %d", s.Window)
	}
}

func TestBenchHandler_NotABenchChecker(t *testing.T) {
	h := NewBenchStatusHandler(&fixedLatencyChecker{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bench", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}
