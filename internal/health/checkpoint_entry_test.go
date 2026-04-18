package health

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckpointHandler_ErrorField(t *testing.T) {
	store := NewCheckpointStore()
	store.Record("svc", Result{Status: StatusUnhealthy, Err: errors.New("timeout")})
	h := NewCheckpointStatusHandler(store)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "timeout") {
		t.Fatalf("expected error in body, got: %s", body)
	}
}

func TestCheckpointHandler_EmptyStore(t *testing.T) {
	h := NewCheckpointStatusHandler(NewCheckpointStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "checkpoints") {
		t.Fatalf("expected checkpoints key")
	}
}

func TestCheckpointStore_Overwrite(t *testing.T) {
	s := NewCheckpointStore()
	s.Record("svc", Result{Status: StatusHealthy})
	s.Record("svc", Result{Status: StatusUnhealthy})
	e, _ := s.Latest("svc")
	if e.Result.Status != StatusUnhealthy {
		t.Fatalf("expected overwrite to unhealthy")
	}
	if len(s.All()) != 1 {
		t.Fatalf("expected single entry after overwrite")
	}
}
