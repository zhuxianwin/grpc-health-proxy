package health

import (
	"context"
	"testing"
)

func makeScoreEntry(service string, weight float64, status Status) scoreEntry {
	c := &stubChecker{result: Result{Status: status}}
	return NewScoreEntry(service, weight, c)
}

type stubChecker struct{ result Result }

func (s *stubChecker) Check(_ context.Context, _ string) Result { return s.result }

func TestScore_AllHealthy(t *testing.T) {
	entries := []scoreEntry{
		makeScoreEntry("a", 1, StatusHealthy),
		makeScoreEntry("b", 1, StatusHealthy),
	}
	c := NewScoreChecker(DefaultScoreConfig(), entries, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestScore_BelowThreshold(t *testing.T) {
	entries := []scoreEntry{
		makeScoreEntry("a", 1, StatusUnhealthy),
		makeScoreEntry("b", 1, StatusUnhealthy),
		makeScoreEntry("c", 1, StatusHealthy),
	}
	cfg := ScoreConfig{Threshold: 0.6}
	c := NewScoreChecker(cfg, entries, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", r.Status)
	}
}

func TestScore_WeightedFavorsHeavy(t *testing.T) {
	entries := []scoreEntry{
		makeScoreEntry("heavy", 10, StatusHealthy),
		makeScoreEntry("light", 1, StatusUnhealthy),
	}
	cfg := ScoreConfig{Threshold: 0.8}
	c := NewScoreChecker(cfg, entries, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy due to heavy weight, got %s", r.Status)
	}
}

func TestScore_EmptyEntries(t *testing.T) {
	c := NewScoreChecker(DefaultScoreConfig(), nil, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("empty entries should be healthy, got %s", r.Status)
	}
}

func TestScore_DefaultConfigOnZeroThreshold(t *testing.T) {
	cfg := ScoreConfig{Threshold: 0}
	entries := []scoreEntry{makeScoreEntry("a", 1, StatusHealthy)}
	c := NewScoreChecker(cfg, entries, nil)
	sc := c.(*scoreChecker)
	if sc.cfg.Threshold != DefaultScoreConfig().Threshold {
		t.Fatalf("expected default threshold %f, got %f", DefaultScoreConfig().Threshold, sc.cfg.Threshold)
	}
}
