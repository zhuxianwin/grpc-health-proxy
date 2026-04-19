package health

import (
	"context"
	"testing"
	"time"
)

func TestStamp_AttachesTimestamp(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	cfg := StampConfig{TimeFunc: func() time.Time { return fixed }}
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewStampChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Metadata["checked_at"] != "2024-01-15T10:00:00Z" {
		t.Fatalf("unexpected checked_at: %q", res.Metadata["checked_at"])
	}
}

func TestStamp_PreservesExistingMetadata(t *testing.T) {
	fixed := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cfg := StampConfig{TimeFunc: func() time.Time { return fixed }}
	inner := &fakeChecker{result: Result{
		Status:   StatusHealthy,
		Metadata: map[string]string{"region": "us-east-1"},
	}}
	c := NewStampChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Metadata["region"] != "us-east-1" {
		t.Fatalf("existing metadata lost")
	}
	if res.Metadata["checked_at"] == "" {
		t.Fatalf("checked_at not set")
	}
}

func TestStamp_DefaultConfigOnZero(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewStampChecker(inner, StampConfig{}, nil)

	res := c.Check(context.Background(), "svc")
	ts, err := time.Parse(time.RFC3339, res.Metadata["checked_at"])
	if err != nil {
		t.Fatalf("invalid timestamp: %v", err)
	}
	if ts.Before(before) {
		t.Fatalf("timestamp too old: %v", ts)
	}
}

func TestStamp_PropagatesUnhealthy(t *testing.T) {
	cfg := DefaultStampConfig()
	inner := &fakeChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStampChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", res.Status)
	}
	if res.Metadata["checked_at"] == "" {
		t.Fatalf("checked_at missing on unhealthy result")
	}
}
