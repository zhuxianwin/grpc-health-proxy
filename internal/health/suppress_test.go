package health

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSuppress_PassesThroughHealthy(t *testing.T) {
	inner := staticChecker(StatusHealthy)
	sc := NewSuppressChecker(inner, DefaultSuppressConfig())
	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestSuppress_SuppressesUnhealthy(t *testing.T) {
	inner := staticChecker(StatusUnhealthy)
	cfg := SuppressConfig{Statuses: []Status{StatusUnhealthy}}
	sc := NewSuppressChecker(inner, cfg)
	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected suppressed to healthy, got %s", res.Status)
	}
	if res.Err != nil {
		t.Fatalf("expected nil err after suppression, got %v", res.Err)
	}
}

func TestSuppress_DoesNotSuppressOtherStatuses(t *testing.T) {
	inner := staticChecker(StatusUnhealthy)
	cfg := SuppressConfig{Statuses: []Status{StatusHealthy}}
	sc := NewSuppressChecker(inner, cfg)
	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
}

func TestSuppress_AddRemoveStatus(t *testing.T) {
	inner := staticChecker(StatusUnhealthy)
	sc := NewSuppressChecker(inner, DefaultSuppressConfig())

	res := sc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatal("expected unhealthy before add")
	}

	sc.AddStatus(StatusUnhealthy)
	res = sc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatal("expected healthy after add")
	}

	sc.RemoveStatus(StatusUnhealthy)
	res = sc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatal("expected unhealthy after remove")
	}
}

func TestSuppressHandler_ReturnsJSON(t *testing.T) {
	inner := staticChecker(StatusHealthy)
	cfg := SuppressConfig{Statuses: []Status{StatusUnhealthy}}
	sc := NewSuppressChecker(inner, cfg)

	h := NewSuppressStatusHandler(sc, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "suppressed_statuses") {
		t.Fatalf("expected suppressed_statuses in body: %s", rr.Body.String())
	}
}

func TestSuppressHandler_NotASuppressChecker(t *testing.T) {
	h := NewSuppressStatusHandler(staticChecker(StatusHealthy), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
