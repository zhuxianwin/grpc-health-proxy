// Package config parses and validates proxy configuration from CLI flags.
package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// Config holds all runtime configuration for the proxy.
type Config struct {
	// HTTPAddr is the address the HTTP server listens on.
	HTTPAddr string
	// GRPCTarget is the gRPC server address to probe.
	GRPCTarget string
	// Services is the list of gRPC service names to check.
	Services []string
	// CacheTTL is how long a health result is cached before re-checking.
	CacheTTL time.Duration
	// WatchInterval is how often the background watcher polls.
	WatchInterval time.Duration
	// RetryConfig controls retry behaviour on transient check failures.
	Retry health.RetryConfig
}

// Parse reads configuration from command-line flags and returns a validated
// Config. It calls flag.Parse internally.
func Parse() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.HTTPAddr, "http-addr", ":8080", "Address for the HTTP health endpoint")
	flag.StringVar(&cfg.GRPCTarget, "grpc-target", "", "gRPC server address (host:port)")
	flag.DurationVar(&cfg.CacheTTL, "cache-ttl", 5*time.Second, "Duration to cache health results")
	flag.DurationVar(&cfg.WatchInterval, "watch-interval", 10*time.Second, "Background watch poll interval")
	flag.IntVar(&cfg.Retry.Attempts, "retry-attempts", 3, "Number of attempts for transient check failures")
	flag.DurationVar(&cfg.Retry.Delay, "retry-delay", 200*time.Millisecond, "Delay between retry attempts")

	flag.Parse()

	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.GRPCTarget == "" {
		return fmt.Errorf("--grpc-target is required")
	}
	if cfg.CacheTTL <= 0 {
		return fmt.Errorf("--cache-ttl must be positive")
	}
	if cfg.WatchInterval <= 0 {
		return fmt.Errorf("--watch-interval must be positive")
	}
	if cfg.Retry.Attempts < 1 {
		return fmt.Errorf("--retry-attempts must be at least 1")
	}
	if cfg.Retry.Delay < 0 {
		return fmt.Errorf("--retry-delay must be non-negative")
	}
	return nil
}
