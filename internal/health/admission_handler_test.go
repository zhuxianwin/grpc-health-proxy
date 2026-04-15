package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdmissionHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	checker := NewAdmissionChecker(inner, AdmissionConfig{MaxConcurrent: 8}, nil)

	h := NewAdmissionStatusHandler(checker, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Active    int64 `json:"active"`
		Max       int64 `json:"max"`
		Available int64 `json:"available"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Max != 8 {
		t.Errorf("expected max=8, got %d", resp.Max)
	}
	if resp.Available != 8 {
		t.Errorf("expected available=8, got %d", resp.Available)
	}
}

func TestAdmissionHandler_ContentType(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	checker := NewAdmissionChecker(inner, DefaultAdmissionConfig(), nil)

	h := NewAdmissionStatusHandler(checker, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestAdmissionHandler_NotAnAdmissionChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewAdmissionStatusHandler(inner, nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}
