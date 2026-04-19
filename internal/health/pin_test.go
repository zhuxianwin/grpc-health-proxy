package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPin_DelegatesWhenUnpinned(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	pc := NewPinChecker(inner, DefaultPinConfig())

	res, err := pc.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", res.Status)
	}
}

func TestPin_ReturnsPinnedResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	pc := NewPinChecker(inner, DefaultPinConfig())
	pc.Pin(Result{Status: StatusUnhealthy})

	res, err := pc.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", res.Status)
	}
}

func TestPin_UnpinRestoresDelegation(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	pc := NewPinChecker(inner, DefaultPinConfig())
	pc.Pin(Result{Status: StatusUnhealthy})
	pc.Unpin()

	res, _ := pc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after unpin, got %v", res.Status)
	}
}

func TestPin_PinnedReturnsNilWhenUnset(t *testing.T) {
	pc := NewPinChecker(&stubChecker{}, DefaultPinConfig())
	if pc.Pinned() != nil {
		t.Fatal("expected nil when not pinned")
	}
}

func TestPinHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	pc := NewPinChecker(inner, DefaultPinConfig())
	pc.Pin(Result{Status: StatusUnhealthy})

	h := NewPinStatusHandler(pc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Pinned bool   `json:"pinned"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.Pinned {
		t.Fatal("expected pinned=true")
	}
	if resp.Status != "unhealthy" {
		t.Fatalf("expected status unhealthy, got %q", resp.Status)
	}
}

func TestPinHandler_NotAPinChecker(t *testing.T) {
	h := NewPinStatusHandler(&stubChecker{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
