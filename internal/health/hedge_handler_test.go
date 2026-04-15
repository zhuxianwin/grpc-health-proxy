package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHedgeHandler_ReturnsJSON(t *testing.T) {
	cfg := HedgeConfig{Delay: 75 * time.Millisecond, MaxHedged: 3}
	h := NewHedgeStatusHandler(cfg, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/hedge", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		DelayMs   int64 `json:"delay_ms"`
		MaxHedged int   `json:"max_hedged"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.DelayMs != 75 {
		t.Errorf("expected delay_ms=75, got %d", body.DelayMs)
	}
	if body.MaxHedged != 3 {
		t.Errorf("expected max_hedged=3, got %d", body.MaxHedged)
	}
}

func TestHedgeHandler_ContentType(t *testing.T) {
	h := NewHedgeStatusHandler(DefaultHedgeConfig(), nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json content-type, got %q", ct)
	}
}

func TestHedgeHandler_DefaultConfig(t *testing.T) {
	def := DefaultHedgeConfig()
	h := NewHedgeStatusHandler(def, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var body struct {
		DelayMs   int64 `json:"delay_ms"`
		MaxHedged int   `json:"max_hedged"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body.DelayMs != def.Delay.Milliseconds() {
		t.Errorf("delay mismatch: want %d got %d", def.Delay.Milliseconds(), body.DelayMs)
	}
	if body.MaxHedged != def.MaxHedged {
		t.Errorf("max_hedged mismatch: want %d got %d", def.MaxHedged, body.MaxHedged)
	}
}
