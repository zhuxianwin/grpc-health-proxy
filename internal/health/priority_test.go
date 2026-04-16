package health

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func makeEntry(name string, c Checker) PriorityEntry {
	return PriorityEntry{Name: name, Checker: c}
}

func TestPriority_FirstSucceeds(t *testing.T) {
	high := &stubChecker{result: Result{Status: StatusHealthy}}
	low := &stubChecker{result: Result{Status: StatusUnhealthy}}

	pc := NewPriorityChecker(PriorityConfig{
		Checkers: []PriorityEntry{makeEntry("high", high), makeEntry("low", low)},
	}, slog.Default())

	res := pc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected Healthy, got %s", res.Status)
	}
	if low.calls != 0 {
		t.Fatal("low priority checker should not have been called")
	}
}

func TestPriority_FallsBackOnError(t *testing.T) {
	high := &stubChecker{result: Result{Status: StatusUnknown, Err: errors.New("dial error")}}
	low := &stubChecker{result: Result{Status: StatusHealthy}}

	pc := NewPriorityChecker(PriorityConfig{
		Checkers: []PriorityEntry{makeEntry("high", high), makeEntry("low", low)},
	}, slog.Default())

	res := pc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected Healthy from fallback, got %s", res.Status)
	}
	if low.calls != 1 {
		t.Fatal("low priority checker should have been called once")
	}
}

func TestPriority_AllFail(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	a := &stubChecker{result: Result{Status: StatusUnknown, Err: err1}}
	b := &stubChecker{result: Result{Status: StatusUnknown, Err: err2}}

	pc := NewPriorityChecker(PriorityConfig{
		Checkers: []PriorityEntry{makeEntry("a", a), makeEntry("b", b)},
	}, nil)

	res := pc.Check(context.Background(), "svc")
	if res.Err == nil {
		t.Fatal("expected an error when all checkers fail")
	}
	if !errors.Is(res.Err, err2) {
		t.Fatalf("expected last error to be wrapped, got: %v", res.Err)
	}
}

func TestPriority_UnhealthyIsConclusive(t *testing.T) {
	// An explicitly unhealthy result (Err == nil) should stop the chain.
	unhealthy := &stubChecker{result: Result{Status: StatusUnhealthy}}
	healthy := &stubChecker{result: Result{Status: StatusHealthy}}

	pc := NewPriorityChecker(PriorityConfig{
		Checkers: []PriorityEntry{makeEntry("primary", unhealthy), makeEntry("backup", healthy)},
	}, slog.Default())

	res := pc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected Unhealthy, got %s", res.Status)
	}
	if healthy.calls != 0 {
		t.Fatal("backup checker should not have been called")
	}
}

func TestPriority_PanicsWithNoCheckers(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic with empty checkers")
		}
	}()
	NewPriorityChecker(PriorityConfig{}, slog.Default())
}
