package health

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestTag_PassesThroughResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	tc := NewTagChecker(inner, DefaultTagConfig(), nil)
	res := tc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestTag_AttachesTags(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := TagConfig{Tags: map[string]string{"env": "prod", "region": "us-east-1"}}
	tc := NewTagChecker(inner, cfg, nil)
	res := tc.Check(context.Background(), "svc")
	if res.Tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", res.Tags["env"])
	}
	if res.Tags["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", res.Tags["region"])
	}
}

func TestTag_MergesWithExistingTags(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, Tags: map[string]string{"source": "inner"}}}
	cfg := TagConfig{Tags: map[string]string{"env": "staging"}}
	tc := NewTagChecker(inner, cfg, nil)
	res := tc.Check(context.Background(), "svc")
	if res.Tags["source"] != "inner" {
		t.Errorf("expected source=inner")
	}
	if res.Tags["env"] != "staging" {
		t.Errorf("expected env=staging")
	}
}

func TestTag_EmptyTagsNoChange(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	tc := NewTagChecker(inner, DefaultTagConfig(), nil)
	res := tc.Check(context.Background(), "svc")
	if len(res.Tags) != 0 {
		t.Errorf("expected no tags")
	}
}

func TestTagHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := TagConfig{Tags: map[string]string{"tier": "frontend"}}
	tc := NewTagChecker(inner, cfg, nil)
	h := NewTagStatusHandler(tc)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	tags, ok := body["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("expected tags map")
	}
	if tags["tier"] != "frontend" {
		t.Errorf("expected tier=frontend")
	}
}

func TestTagHandler_NotATagChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewTagStatusHandler(inner)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 400 {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
