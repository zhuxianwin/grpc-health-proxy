package health

import (
	"context"
	"testing"
)

func TestMask_PassesThroughWhenNoKeys(t *testing.T) {
	inner := &fakeChecker{result: Result{
		Status:   StatusHealthy,
		Metadata: map[string]string{"token": "secret", "env": "prod"},
	}}
	c := NewMaskChecker(inner, DefaultMaskConfig())
	res := c.Check(context.Background(), "svc")
	if res.Metadata["token"] != "secret" {
		t.Errorf("expected token to be unmasked, got %q", res.Metadata["token"])
	}
}

func TestMask_RedactsListedKeys(t *testing.T) {
	inner := &fakeChecker{result: Result{
		Status:   StatusHealthy,
		Metadata: map[string]string{"token": "secret", "env": "prod"},
	}}
	cfg := DefaultMaskConfig()
	cfg.Keys = []string{"token"}
	c := NewMaskChecker(inner, cfg)
	res := c.Check(context.Background(), "svc")
	if res.Metadata["token"] != "***" {
		t.Errorf("expected token to be masked, got %q", res.Metadata["token"])
	}
	if res.Metadata["env"] != "prod" {
		t.Errorf("expected env to be unmasked, got %q", res.Metadata["env"])
	}
}

func TestMask_MultipleKeys(t *testing.T) {
	inner := &fakeChecker{result: Result{
		Status:   StatusHealthy,
		Metadata: map[string]string{"token": "abc", "password": "xyz", "region": "us-east-1"},
	}}
	cfg := DefaultMaskConfig()
	cfg.Keys = []string{"token", "password"}
	c := NewMaskChecker(inner, cfg)
	res := c.Check(context.Background(), "svc")
	if res.Metadata["token"] != "***" || res.Metadata["password"] != "***" {
		t.Error("expected both sensitive keys to be masked")
	}
	if res.Metadata["region"] != "us-east-1" {
		t.Errorf("unexpected region value: %q", res.Metadata["region"])
	}
}

func TestMask_EmptyMetadataNoOp(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultMaskConfig()
	cfg.Keys = []string{"token"}
	c := NewMaskChecker(inner, cfg)
	res := c.Check(context.Background(), "svc")
	if len(res.Metadata) != 0 {
		t.Errorf("expected empty metadata, got %v", res.Metadata)
	}
}

func TestMask_NilLoggerUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := MaskConfig{Keys: []string{"k"}}
	c := NewMaskChecker(inner, cfg)
	if c == nil {
		t.Fatal("expected non-nil checker")
	}
}
