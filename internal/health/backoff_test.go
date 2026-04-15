package health

import (
	"testing"
	"time"
)

func TestDefaultBackoffConfig(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if cfg.InitialInterval != 100*time.Millisecond {
		t.Fatalf("expected 100ms initial interval, got %v", cfg.InitialInterval)
	}
	if cfg.Multiplier != 2.0 {
		t.Fatalf("expected multiplier 2.0, got %v", cfg.Multiplier)
	}
}

func TestBackoff_FirstAttempt(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     10 * time.Second,
		Jitter:          0,
	}
	got := cfg.Backoff(0, 0)
	if got != 100*time.Millisecond {
		t.Fatalf("attempt 0: expected 100ms, got %v", got)
	}
}

func TestBackoff_Exponential(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     10 * time.Second,
		Jitter:          0,
	}
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		got := cfg.Backoff(i, 0)
		if got != want {
			t.Errorf("attempt %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestBackoff_CappedAtMaxInterval(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     500 * time.Millisecond,
		Jitter:          0,
	}
	got := cfg.Backoff(10, 0)
	if got != 500*time.Millisecond {
		t.Fatalf("expected cap at 500ms, got %v", got)
	}
}

func TestBackoff_Jitter(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      1.0, // no growth, easier to reason about
		MaxInterval:     1 * time.Second,
		Jitter:          0.5,
	}
	// randFrac=1 means full jitter applied: 100ms + 50% = 150ms
	got := cfg.Backoff(0, 1.0)
	if got != 150*time.Millisecond {
		t.Fatalf("expected 150ms with full jitter, got %v", got)
	}
	// randFrac=0 means no jitter: exactly 100ms
	got = cfg.Backoff(0, 0)
	if got != 100*time.Millisecond {
		t.Fatalf("expected 100ms with zero jitter, got %v", got)
	}
}

func TestBackoff_ZeroInitialInterval(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 0,
		Multiplier:      2.0,
	}
	got := cfg.Backoff(5, 0.5)
	if got != 0 {
		t.Fatalf("expected 0 when InitialInterval is 0, got %v", got)
	}
}
