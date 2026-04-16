package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"context"
)

func TestQuotaHandler_ReturnsJSON(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, QuotaConfig{MaxCalls: 10, Window: time.Minute}, := NewQuotaStatusHandler(qc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/quota	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["max_calls"]; !ok {
		t.Error("missing max_calls field")
	}
}

func TestQuotaHandler_ExhaustedTrue(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, QuotaConfig{MaxCalls: 2, Window: time.Minute}, nil)

	qc.Check(context.Background(), "svc")
	qc.Check(context.Background(), "svc")
	qc.Check(context.Background(), "svc") // exceeds

	h := NewQuotaStatusHandler(qc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/quota", nil))

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if exhausted, _ := resp["exhausted"].(bool); !exhausted {
		t.Error("expected exhausted=true")
	}
}

func TestQuotaHandler_NotAQuotaChecker(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	h := NewQuotaStatusHandler(inner, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/quota", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
