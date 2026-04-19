package health

import (
	"context"
	"log/slog"
	"sync"
)

// MirrorConfig controls the mirror checker behaviour.
type MirrorConfig struct {
	// Logger is used for mirror errors; defaults to slog.Default().
	Logger *slog.Logger
}

// DefaultMirrorConfig returns sensible defaults.
func DefaultMirrorConfig() MirrorConfig {
	return MirrorConfig{Logger: slog.Default()}
}

// mirrorChecker sends every check to a primary and a set of mirror targets.
// The primary result is always returned; mirrors are fired asynchronously.
type mirrorChecker struct {
	primary Checker
	mirrors []Checker
	cfg     MirrorConfig
}

// NewMirrorChecker wraps primary and fires each check against mirrors in the
// background. Mirror results are discarded; only the primary result matters.
func NewMirrorChecker(primary Checker, mirrors []Checker, cfg MirrorConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &mirrorChecker{primary: primary, mirrors: mirrors, cfg: cfg}
}

func (m *mirrorChecker) Check(ctx context.Context, service string) Result {
	result := m.primary.Check(ctx, service)

	var wg sync.WaitGroup
	for _, mirror := range m.mirrors {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			r := c.Check(ctx, service)
			if r.Err != nil {
				m.cfg.Logger.Warn("mirror check error",
					"service", service,
					"err", r.Err)
			}
		}(mirror)
	}
	wg.Wait()

	return result
}
