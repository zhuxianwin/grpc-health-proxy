package config

import (
	"flag"
	"fmt"
	"time"
)

// Config holds all runtime configuration for the proxy.
type Config struct {
	// GRPCAddr is the address of the upstream gRPC health endpoint.
	GRPCAddr string
	// HTTPPort is the port on which the HTTP server listens.
	HTTPPort int
	// ServiceName is the gRPC service name to check (empty means overall server health).
	ServiceName string
	// DialTimeout is the maximum duration to wait when connecting to the gRPC server.
	DialTimeout time.Duration
	// CheckTimeout is the maximum duration to wait for a health check response.
	CheckTimeout time.Duration
	// TLSEnabled controls whether TLS is used when connecting to the gRPC server.
	TLSEnabled bool
}

// Parse reads configuration from command-line flags and returns a validated Config.
func Parse() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.GRPCAddr, "grpc-addr", "localhost:50051", "Address of the upstream gRPC server (host:port)")
	flag.IntVar(&cfg.HTTPPort, "http-port", 8086, "Port for the HTTP health endpoint")
	flag.StringVar(&cfg.ServiceName, "service", "", "gRPC service name to probe (empty = server-level check)")
	flag.DurationVar(&cfg.DialTimeout, "dial-timeout", 5*time.Second, "Timeout for dialing the gRPC server")
	flag.DurationVar(&cfg.CheckTimeout, "check-timeout", 3*time.Second, "Timeout for the health check RPC")
	flag.BoolVar(&cfg.TLSEnabled, "tls", false, "Use TLS when connecting to the gRPC server")
	flag.Parse()

	if cfg.GRPCAddr == "" {
		return nil, fmt.Errorf("grpc-addr must not be empty")
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return nil, fmt.Errorf("http-port %d is out of valid range", cfg.HTTPPort)
	}

	return cfg, nil
}
