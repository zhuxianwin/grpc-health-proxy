package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultPeakConfig returns a PeakConfig with sensible defaults.
func DefaultPeakConfig() PeakConfig {
	return PeakConfig{
		Window: 60 * time.Second,
	}
}

// PeakConfig controls the behaviour of NewPeakChecker.
type PeakConfig struct {
	// Window is the rolling duration over which peak latency is tracked.
	Window time.Duration
}

type peakEntry struct {
	latency time.Duration
	at      time.Time
}

// peakChecker records the highest observed check latency within a
// sliding window and exposes it via the Result metadata.
type peakChecker struct {
	inner  Checker
	cfg    PeakConfig
	log    *slog.Logger
	mu     sync.Mutex
	entries []peakEntry
}

// NewPeakChecker wraps inner and annotates each Result with the peak
// (maximum) latency observed within the configured Window.
func NewPeakChecker(inner Checker, cfg PeakConfig, log *slog.Logger) Checker {
	if cfg.Window <= 0 {
		cfg = DefaultPeakConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &peakChecker{inner: inner, cfg: cfg, log: log}
}

func (p *peakChecker) Check(ctx context.Context, service string) Result {
	start := time.Now()
	res := p.inner.Check(ctx, service)
	elapsed := time.Since(start)

	now := time.Now()
	p.mu.Lock()
	p.evict(now)
	p.entries = append(p.entries, peakEntry{latency: elapsed, at: now})
	peak := p.peak()
	p.mu.Unlock()

	if res.Metadata == nil {
		res.Metadata = make(map[string]string)
	}
	res.Metadata["peak_latency_ms"] = formatDurationMS(peak)
	res.Metadata["last_latency_ms"] = formatDurationMS(elapsed)
	return res
}

// evict removes entries older than the window. Must be called with p.mu held.
func (p *peakChecker) evict(now time.Time) {
	cutoff := now.Add(-p.cfg.Window)
	i := 0
	for i < len(p.entries) && p.entries[i].at.Before(cutoff) {
		i++
	}
	p.entries = p.entries[i:]
}

// peak returns the maximum latency in the current window. Must be called with p.mu held.
func (p *peakChecker) peak() time.Duration {
	var max time.Duration
	for _, e := range p.entries {
		if e.latency > max {
			max = e.latency
		}
	}
	return max
}

func formatDurationMS(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 0 {
		return "0"
	}
	return itoa(int(ms))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
