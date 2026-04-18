package health

import (
	"context"
	"testing"
)

type fixedChecker struct{ result Result }

func (f *fixedChecker) Check(_ context.Context, _ string) Result { return f.result }

func TestLabel_PassesThroughResult(t *testing.T) {
	inner := &fixedChecker{result: Healthy("svc")}
	cfg := LabelConfig{Labels: map[string]string{"env": "prod"}}
	c := NewLabelChecker(inner, cfg, nil)
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestLabel_LabelsReturned(t *testing.T) {
	inner := &fixedChecker{result: Healthy("svc")}
	cfg := LabelConfig{Labels: map[string]string{"region": "us-east-1", "env": "staging"}}
	c := NewLabelChecker(inner, cfg, nil)
	lc := c.(*labelChecker)
	labels := lc.Labels()
	if labels["region"] != "us-east-1" {
		t.Fatalf("expected us-east-1, got %s", labels["region"])
	}
	if labels["env"] != "staging" {
		t.Fatalf("expected staging, got %s", labels["env"])
	}
}

func TestLabel_EmptyLabels(t *testing.T) {
	inner := &fixedChecker{result: Unhealthy("svc", nil)}
	cfg := DefaultLabelConfig()
	c := NewLabelChecker(inner, cfg, nil)
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
}

func TestLabel_LabelsCopiedOnCreate(t *testing.T) {
	src := map[string]string{"k": "v"}
	cfg := LabelConfig{Labels: src}
	c := NewLabelChecker(&fixedChecker{result: Healthy("svc")}, cfg, nil)
	src["k"] = "mutated"
	lc := c.(*labelChecker)
	if lc.labels["k"] != "v" {
		t.Fatal("labels should be copied on construction")
	}
}

func TestDefaultLabelConfig(t *testing.T) {
	cfg := DefaultLabelConfig()
	if cfg.Labels == nil {
		t.Fatal("expected non-nil labels map")
	}
	if len(cfg.Labels) != 0 {
		t.Fatalf("expected empty labels, got %d", len(cfg.Labels))
	}
}
