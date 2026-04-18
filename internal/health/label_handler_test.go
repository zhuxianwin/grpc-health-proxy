package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabelHandler_ReturnsJSON(t *testing.T) {
	inner := &fixedChecker{result: Healthy("svc")}
	cfg := LabelConfig{Labels: map[string]string{"env": "prod", "team": "platform"}}
	c := NewLabelChecker(inner, cfg, nil)
	h := NewLabelStatusHandler(c, nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp labelStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Labels["env"] != "prod" {
		t.Fatalf("expected prod, got %s", resp.Labels["env"])
	}
}

func TestLabelHandler_ContentType(t *testing.T) {
	inner := &fixedChecker{result: Healthy("svc")}
	c := NewLabelChecker(inner, DefaultLabelConfig(), nil)
	h := NewLabelStatusHandler(c, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestLabelHandler_NotALabelChecker(t *testing.T) {
	inner := &fixedChecker{result: Healthy("svc")}
	h := NewLabelStatusHandler(inner, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
