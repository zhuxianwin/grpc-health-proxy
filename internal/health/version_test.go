package health

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestVersion_PassesThroughNoConstraint(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy, Metadata: map[string]string{"version": "1.2.3"}}
	})
	vc := NewVersionChecker(inner, VersionConfig{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	r := vc.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestVersion_AboveMinVersionHealthy(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy, Metadata: map[string]string{"version": "2.0.0"}}
	})
	cfg := VersionConfig{MinVersion: "1.0.0"}
	vc := NewVersionChecker(inner, cfg, nil)
	r := vc.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestVersion_BelowMinVersionUnhealthy(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy, Metadata: map[string]string{"version": "0.9.0"}}
	})
	cfg := VersionConfig{MinVersion: "1.0.0"}
	vc := NewVersionChecker(inner, cfg, nil)
	r := vc.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", r.Status)
	}
	if r.Err != ErrVersionTooLow {
		t.Fatalf("expected ErrVersionTooLow, got %v", r.Err)
	}
}

func TestVersion_NoVersionMetaSkipsConstraint(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy}
	})
	cfg := VersionConfig{MinVersion: "1.0.0"}
	vc := NewVersionChecker(inner, cfg, nil)
	r := vc.Check(context.Background(), "svc")
	// No version key — constraint not applied, result passes through.
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy when no version metadata, got %s", r.Status)
	}
}

func TestVersion_DefaultConfigOnZero(t *testing.T) {
	cfg := DefaultVersionConfig()
	if cfg.Timeout == 0 {
		t.Fatal("expected non-zero default timeout")
	}
}
