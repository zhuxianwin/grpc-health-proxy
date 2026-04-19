package health

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalize_TrimSpace(t *testing.T) {
	var got string
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		got = svc
		return Healthy(svc)
	})
	c := NewNormalizeChecker(inner, DefaultNormalizeConfig())
	c.Check(context.Background(), "  myservice  ")
	if got != "myservice" {
		t.Fatalf("expected trimmed service, got %q", got)
	}
}

func TestNormalize_ToLower(t *testing.T) {
	var got string
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		got = svc
		return Healthy(svc)
	})
	cfg := NormalizeConfig{TrimSpace: false, ToLower: true}
	c := NewNormalizeChecker(inner, cfg)
	c.Check(context.Background(), "MyService")
	if got != "myservice" {
		t.Fatalf("expected lowercase service, got %q", got)
	}
}

func TestNormalize_NoChange(t *testing.T) {
	var got string
	inner := CheckerFunc(func(_ context.Context, svc string) Result {
		got = svc
		return Healthy(svc)
	})
	cfg := NormalizeConfig{TrimSpace: false, ToLower: false}
	c := NewNormalizeChecker(inner, cfg)
	c.Check(context.Background(), "MyService")
	if got != "MyService" {
		t.Fatalf("expected unchanged service, got %q", got)
	}
}

func TestNormalizeHandler_ReturnsJSON(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, svc string) Result { return Healthy(svc) })
	c := NewNormalizeChecker(inner, DefaultNormalizeConfig())
	h := NewNormalizeStatusHandler(c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "trim_space") {
		t.Fatalf("expected trim_space in body, got %s", rr.Body.String())
	}
}

func TestNormalizeHandler_NotANormalizeChecker(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, svc string) Result { return Healthy(svc) })
	h := NewNormalizeStatusHandler(inner)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
