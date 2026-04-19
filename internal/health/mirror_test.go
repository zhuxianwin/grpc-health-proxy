package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type countingChecker struct {
	calls atomic.Int32
	result Result
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	c.calls.Add(1)
	return c.result
}

func TestMirror_PrimaryResultReturned(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	mirror := &countingChecker{result: Result{Status: StatusUnhealthy}}

	mc := NewMirrorChecker(primary, []Checker{mirror}, DefaultMirrorConfig())
	res := mc.Check(context.Background(), "svc")

	if res.Status != StatusHealthy {
		t.Fatalf("expected primary result, got %v", res.Status)
	}
}

func TestMirror_MirrorsAreCalled(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	m1 := &countingChecker{result: Result{Status: StatusHealthy}}
	m2 := &countingChecker{result: Result{Status: StatusHealthy}}

	mc := NewMirrorChecker(primary, []Checker{m1, m2}, DefaultMirrorConfig())
	_ = mc.Check(context.Background(), "svc")

	if m1.calls.Load() != 1 || m2.calls.Load() != 1 {
		t.Fatal("expected both mirrors to be called once")
	}
}

func TestMirror_MirrorErrorDoesNotPropagate(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	errMirror := &countingChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("boom")}}

	mc := NewMirrorChecker(primary, []Checker{errMirror}, DefaultMirrorConfig())
	res := mc.Check(context.Background(), "svc")

	if res.Err != nil {
		t.Fatalf("mirror error leaked into primary result: %v", res.Err)
	}
}

func TestMirror_NoMirrors(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	mc := NewMirrorChecker(primary, nil, DefaultMirrorConfig())
	res := mc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status: %v", res.Status)
	}
}

func TestMirrorHandler_ReturnsJSON(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	m1 := &countingChecker{}
	m2 := &countingChecker{}
	mc := NewMirrorChecker(primary, []Checker{m1, m2}, DefaultMirrorConfig())

	h := NewMirrorStatusHandler(mc)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if body == "" {
		t.Fatal("empty body")
	}
}

func TestMirrorHandler_NotAMirrorChecker(t *testing.T) {
	h := NewMirrorStatusHandler(&countingChecker{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
