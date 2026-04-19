package health

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRelay_PrimaryResultReturned(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	var got Result
	var wg sync.WaitGroup
	wg.Add(1)
	cfg := RelayConfig{
		Sink: func(_ string, r Result) {
			got = r
			wg.Done()
		},
	}
	c := NewRelayChecker(inner, cfg)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", r.Status)
	}
	wg.Wait()
	if got.Status != StatusHealthy {
		t.Fatalf("sink got wrong status: %v", got.Status)
	}
}

func TestRelay_SinkReceivesServiceName(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	var gotSvc string
	var wg sync.WaitGroup
	wg.Add(1)
	cfg := RelayConfig{
		Sink: func(svc string, _ Result) {
			gotSvc = svc
			wg.Done()
		},
	}
	c := NewRelayChecker(inner, cfg)
	c.Check(context.Background(), "my-service")
	wg.Wait()
	if gotSvc != "my-service" {
		t.Fatalf("expected my-service, got %q", gotSvc)
	}
}

func TestRelay_NilSinkUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewRelayChecker(inner, RelayConfig{Sink: nil})
	// Should not panic
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("unexpected status: %v", r.Status)
	}
}

func TestRelay_PanicInSinkDoesNotPropagate(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := RelayConfig{
		Sink: func(_ string, _ Result) { panic("boom") },
	}
	c := NewRelayChecker(inner, cfg)
	r := c.Check(context.Background(), "svc")
	// Give goroutine time to panic and recover
	time.Sleep(20 * time.Millisecond)
	if r.Status != StatusHealthy {
		t.Fatalf("unexpected status: %v", r.Status)
	}
}

func TestDefaultRelayConfig(t *testing.T) {
	cfg := DefaultRelayConfig()
	if cfg.Sink == nil {
		t.Fatal("expected non-nil sink")
	}
	if cfg.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
}
