package health

import (
	"context"
	"log/slog"
	"sync"
)

// ScoreConfig holds configuration for the score-based health checker.
type ScoreConfig struct {
	// Threshold is the minimum score [0.0, 1.0] to be considered healthy.
	Threshold float64
	// Weights maps service names to their relative weight.
	Weights map[string]float64
}

// DefaultScoreConfig returns a ScoreConfig with sensible defaults.
func DefaultScoreConfig() ScoreConfig {
	return ScoreConfig{
		Threshold: 0.5,
		Weights:   map[string]float64{},
	}
}

// scoreChecker aggregates multiple checkers into a weighted health score.
type scoreChecker struct {
	cfg      ScoreConfig
	entries  []scoreEntry
	mu       sync.Mutex
	logger   *slog.Logger
}

type scoreEntry struct {
	service string
	weight  float64
	checker Checker
}

// NewScoreChecker returns a Checker that computes a weighted score across
// multiple checkers. It is healthy when the weighted score meets the threshold.
func NewScoreChecker(cfg ScoreConfig, entries []scoreEntry, logger *slog.Logger) Checker {
	if cfg.Threshold <= 0 {
		cfg.Threshold = DefaultScoreConfig().Threshold
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &scoreChecker{cfg: cfg, entries: entries, logger: logger}
}

// NewScoreEntry creates a scoreEntry for use with NewScoreChecker.
func NewScoreEntry(service string, weight float64, c Checker) scoreEntry {
	if weight <= 0 {
		weight = 1.0
	}
	return scoreEntry{service: service, weight: weight, checker: c}
}

func (s *scoreChecker) Check(ctx context.Context, service string) Result {
	s.mu.Lock()
	entries := s.entries
	s.mu.Unlock()

	var totalWeight, scoreSum float64
	for _, e := range entries {
		r := e.checker.Check(ctx, e.service)
		totalWeight += e.weight
		if r.Status == StatusHealthy {
			scoreSum += e.weight
		}
	}

	if totalWeight == 0 {
		return Result{Status: StatusHealthy}
	}

	score := scoreSum / totalWeight
	s.logger.Debug("score checker", "service", service, "score", score, "threshold", s.cfg.Threshold)

	if score >= s.cfg.Threshold {
		return Result{Status: StatusHealthy, Metadata: map[string]string{"score": formatScore(score)}}
	}
	return Result{Status: StatusUnhealthy, Metadata: map[string]string{"score": formatScore(score)}}
}

func formatScore(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
