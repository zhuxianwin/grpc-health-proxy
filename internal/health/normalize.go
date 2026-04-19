package health

import (
	"context"
	"log/slog"
	"strings"
)

// NormalizeConfig holds options for the NormalizeChecker.
type NormalizeConfig struct {
	// TrimSpace removes leading/trailing whitespace from the service name.
	TrimSpace bool
	// ToLower lowercases the service name before delegating.
	ToLower bool
	// Logger is an optional structured logger.
	Logger *slog.Logger
}

// DefaultNormalizeConfig returns sensible defaults.
func DefaultNormalizeConfig() NormalizeConfig {
	return NormalizeConfig{
		TrimSpace: true,
		ToLower:   false,
		Logger:    slog.Default(),
	}
}

// normalizeChecker normalises the service name before forwarding to the inner checker.
type normalizeChecker struct {
	cfg   NormalizeConfig
	inner Checker
}

// NewNormalizeChecker wraps inner, normalising the service name on each call.
func NewNormalizeChecker(inner Checker, cfg NormalizeConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &normalizeChecker{cfg: cfg, inner: inner}
}

func (n *normalizeChecker) Check(ctx context.Context, service string) Result {
	normalized := service
	if n.cfg.TrimSpace {
		normalized = strings.TrimSpace(normalized)
	}
	if n.cfg.ToLower {
		normalized = strings.ToLower(normalized)
	}
	if normalized != service {
		n.cfg.Logger.Debug("normalized service name", "original", service, "normalized", normalized)
	}
	return n.inner.Check(ctx, normalized)
}
