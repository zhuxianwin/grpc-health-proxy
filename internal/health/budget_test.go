package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBudget_AllowsWhenUnderThreshold(t *testing.T) {
	cfg := DefaultBudgetConfig()
	cfg.MinSamples = 3
	cfg.ErrorBudget = 0.5 // 50%

	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	bc := NewBudgetChecker(inner, cfg)

	for i := 0; i < 3; i++ {
		res := bc.Check(context.Background(), "svc")
		if res.Status != StatusHealthy {
			t.Fatalf("expected healthy, got %s", res.Status)
		}
	}
}

func TestBudget_ExhaustsAfterTooManyErrors(t *testing.T) {
	cfg := DefaultBudgetConfig()
	cfg.MinSamples = 2
	cfg.ErrorBudget = 0.4 // 40%

	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		if calls%2 == 0 {
			return Result{Status: StatusUnhealthy, Err: errors.New("boom")}
		}
		return Result{Status: StatusUnhealthy, Err: errors.New("boom")}
	})
	bc := NewBudgetChecker(inner, cfg)

	bc.Check(context.Background(), "svc") // fail
	bc.Check(context.Background(), "svc") // fail — 100% > 40%
	res := bc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
	if !errors.Is(res.Err, ErrBudgetExhausted) {
		t.Fatalf("expected ErrBudgetExhausted, got %v", res.Err)
	}
}

func TestBudget_BelowMinSamplesPassesThrough(t *testing.T) {
	cfg := DefaultBudgetConfig()
	cfg.MinSamples = 10

	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	bc := NewBudgetChecker(inner, cfg)

	res := bc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected inner unhealthy, got %s", res.Status)
	}
	if errors.Is(res.Err, ErrBudgetExhausted) {
		t.Fatal("should not be ErrBudgetExhausted below min samples")
	}
}

func TestBudget_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	bc := NewBudgetChecker(inner, BudgetConfig{})
	if bc.cfg.Window != DefaultBudgetConfig().Window {
		t.Fatalf("expected default window, got %s", bc.cfg.Window)
	}
}

func TestBudgetHandler_ReturnsJSON(t *testing.T) {
	cfg := DefaultBudgetConfig()
	cfg.MinSamples = 1
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	bc := NewBudgetChecker(inner, cfg)
	bc.Check(context.Background(), "mysvc")

	h := NewBudgetStatusHandler(bc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?service=mysvc", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "mysvc") {
		t.Fatalf("body missing service name: %s", rec.Body.String())
	}
}

func TestBudgetHandler_MissingServiceParam(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	bc := NewBudgetChecker(inner, DefaultBudgetConfig())
	h := NewBudgetStatusHandler(bc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// checkerFunc is a function adapter for the Checker interface.
type checkerFunc func(ctx context.Context, service string) Result

func (f checkerFunc) Check(ctx context.Context, service string) Result { return f(ctx, service) }

// stubChecker returns a fixed result.
type stubChecker struct{ result Result }

func (s *stubChecker) Check(_ context.Context, _ string) Result { return s.result }

var _ Checker = (*stubChecker)(nil)
var _ Checker = (checkerFunc)(nil)

// ensure window eviction does not panic when entries age out
func TestBudget_WindowEviction(t *testing.T) {
	cfg := DefaultBudgetConfig()
	cfg.Window = 1 * time.Millisecond
	cfg.MinSamples = 1

	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	bc := NewBudgetChecker(inner, cfg)
	bc.Check(context.Background(), "svc")
	time.Sleep(5 * time.Millisecond)
	bc.Check(context.Background(), "svc") // old entry should be evicted
	total, _ := bc.BudgetStats("svc")
	if total != 1 {
		t.Fatalf("expected 1 entry after eviction, got %d", total)
	}
}
