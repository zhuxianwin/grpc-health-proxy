package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fixedChecker struct {
	result Result
	delay  time.Duration
}

func (f *fixedChecker) Check(ctx context.Context, _ string) Result {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return Result{Status: StatusUnhealthy, Err: ctx.Err()}
		}
	}
	return f.result
}

func TestBatch_AllHealthy(t *testing.T) {
	checkers := map[string]Checker{
		"a": &fixedChecker{result: Result{Status: StatusHealthy}},
		"b": &fixedChecker{result: Result{Status: StatusHealthy}},
	}
	bc := NewBatchChecker(checkers, DefaultBatchConfig())
	results := bc.CheckAll(context.Background())
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Result.Status != StatusHealthy {
			t.Errorf("%s: expected healthy", r.Target)
		}
	}
}

func TestBatch_OneUnhealthy(t *testing.T) {
	checkers := map[string]Checker{
		"ok":  &fixedChecker{result: Result{Status: StatusHealthy}},
		"bad": &fixedChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}},
	}
	bc := NewBatchChecker(checkers, DefaultBatchConfig())
	results := bc.CheckAll(context.Background())
	unhealthy := 0
	for _, r := range results {
		if r.Result.Status == StatusUnhealthy {
			unhealthy++
		}
	}
	if unhealthy != 1 {
		t.Errorf("expected 1 unhealthy, got %d", unhealthy)
	}
}

func TestBatch_ConcurrencyLimit(t *testing.T) {
	checkers := map[string]Checker{}
	for i := 0; i < 10; i++ {
		checkers[string(rune('a'+i))] = &fixedChecker{
			result: Result{Status: StatusHealthy},
			delay:  5 * time.Millisecond,
		}
	}
	cfg := BatchConfig{MaxConcurrency: 3}
	bc := NewBatchChecker(checkers, cfg)
	start := time.Now()
	results := bc.CheckAll(context.Background())
	if time.Since(start) < 5*time.Millisecond {
		t.Error("expected some serialisation delay")
	}
	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}
}

func TestBatch_DefaultConfigOnZero(t *testing.T) {
	bc := NewBatchChecker(map[string]Checker{}, BatchConfig{})
	if cap(bc.sem) != DefaultBatchConfig().MaxConcurrency {
		t.Errorf("expected default concurrency %d", DefaultBatchConfig().MaxConcurrency)
	}
}
