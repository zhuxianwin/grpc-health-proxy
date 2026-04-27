package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShelf_ServesInnerOnMiss(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		calls++
		return Result{Status: StatusHealthy}
	})
	sc := NewShelfChecker(inner, ShelfConfig{TTL: time.Second})
	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
	if calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", calls)
	}
}

func TestShelf_CachesHealthyResult(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		calls++
		return Result{Status: StatusHealthy}
	})
	sc := NewShelfChecker(inner, ShelfConfig{TTL: time.Minute})
	sc.Check(context.Background(), "svc")
	sc.Check(context.Background(), "svc")
	if calls != 1 {
		t.Fatalf("expected 1 inner call after cache hit, got %d", calls)
	}
}

func TestShelf_DoesNotCacheUnhealthy(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	sc := NewShelfChecker(inner, ShelfConfig{TTL: time.Minute})
	sc.Check(context.Background(), "svc")
	sc.Check(context.Background(), "svc")
	if calls != 2 {
		t.Fatalf("expected 2 inner calls for unhealthy, got %d", calls)
	}
}

func TestShelf_EvictsExpiredEntry(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		calls++
		return Result{Status: StatusHealthy}
	})
	sc := NewShelfChecker(inner, ShelfConfig{TTL: 10 * time.Millisecond})
	sc.Check(context.Background(), "svc")
	time.Sleep(20 * time.Millisecond)
	sc.Check(context.Background(), "svc")
	if calls != 2 {
		t.Fatalf("expected 2 inner calls after expiry, got %d", calls)
	}
}

func TestShelf_DefaultConfigOnZero(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy}
	})
	// Zero TTL should fall back to default without panicking.
	sc := NewShelfChecker(inner, ShelfConfig{})
	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %s", res.Status)
	}
}

func TestShelfHandler_ReturnsJSON(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy}
	})
	sc := NewShelfChecker(inner, ShelfConfig{TTL: time.Minute})
	sc.Check(context.Background(), "alpha")

	h := NewShelfStatusHandler(sc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload []shelfHandlerPayload
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(payload) != 1 || payload[0].Service != "alpha" || !payload[0].Shelved {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestShelfHandler_NotAShelfChecker(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		return Result{Status: StatusHealthy}
	})
	h := NewShelfStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
