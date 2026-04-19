package health

import (
	"context"
	"testing"
)

func TestMetadata_AttachesStaticMeta(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	cfg := MetadataConfig{Meta: map[string]string{"env": "prod", "region": "us-east-1"}}
	c := NewMetadataChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Metadata["env"] != "prod" {
		t.Fatalf("expected env=prod, got %q", res.Metadata["env"])
	}
	if res.Metadata["region"] != "us-east-1" {
		t.Fatalf("expected region=us-east-1, got %q", res.Metadata["region"])
	}
}

func TestMetadata_ResultMetaOverridesStatic(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{
			Status:   StatusHealthy,
			Metadata: map[string]string{"env": "staging"},
		}
	})
	cfg := MetadataConfig{Meta: map[string]string{"env": "prod"}}
	c := NewMetadataChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Metadata["env"] != "staging" {
		t.Fatalf("result metadata should take precedence, got %q", res.Metadata["env"])
	}
}

func TestMetadata_EmptyConfigNoChange(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	c := NewMetadataChecker(inner, DefaultMetadataConfig(), nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %v", res.Status)
	}
}

func TestMetadata_PropagatesUnhealthy(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusUnhealthy}
	})
	cfg := MetadataConfig{Meta: map[string]string{"team": "platform"}}
	c := NewMetadataChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", res.Status)
	}
	if res.Metadata["team"] != "platform" {
		t.Fatalf("expected team=platform, got %q", res.Metadata["team"])
	}
}

func TestDefaultMetadataConfig(t *testing.T) {
	cfg := DefaultMetadataConfig()
	if cfg.Meta == nil {
		t.Fatal("expected non-nil Meta map")
	}
	if len(cfg.Meta) != 0 {
		t.Fatalf("expected empty Meta, got %d entries", len(cfg.Meta))
	}
}
