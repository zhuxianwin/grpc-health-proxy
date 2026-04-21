package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFlipHandler_ToggleViaPost(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	h := NewFlipStatusHandler(fc)

	req := httptest.NewRequest(http.MethodPost, "/?action=toggle", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "true") {
		t.Fatalf("expected active:true in body, got %s", body)
	}
	if !fc.Active() {
		t.Fatal("expected flip to be active after POST toggle")
	}
}

func TestFlipHandler_GetReportsState(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	h := NewFlipStatusHandler(fc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "false") {
		t.Fatalf("expected active:false in body, got %s", body)
	}
}

func TestFlipHandler_ContentType(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	h := NewFlipStatusHandler(fc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestFlipHandler_ToggleTwiceRestoresState(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	h := NewFlipStatusHandler(fc)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/?action=toggle", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("toggle %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	if fc.Active() {
		t.Fatal("expected flip to be inactive after two toggles")
	}
}
