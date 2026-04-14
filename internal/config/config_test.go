package config

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := &Config{
		GRPCAddr:     "localhost:50051",
		HTTPPort:     8086,
		ServiceName:  "",
		DialTimeout:  5 * time.Second,
		CheckTimeout: 3 * time.Second,
		TLSEnabled:   false,
	}

	if cfg.GRPCAddr != "localhost:50051" {
		t.Errorf("expected default grpc-addr localhost:50051, got %s", cfg.GRPCAddr)
	}
	if cfg.HTTPPort != 8086 {
		t.Errorf("expected default http-port 8086, got %d", cfg.HTTPPort)
	}
	if cfg.DialTimeout != 5*time.Second {
		t.Errorf("expected dial timeout 5s, got %v", cfg.DialTimeout)
	}
	if cfg.CheckTimeout != 3*time.Second {
		t.Errorf("expected check timeout 3s, got %v", cfg.CheckTimeout)
	}
	if cfg.TLSEnabled {
		t.Error("expected TLS disabled by default")
	}
}

func validate(cfg *Config) error {
	if cfg.GRPCAddr == "" {
		return fmt.Errorf("grpc-addr must not be empty")
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("http-port %d is out of valid range", cfg.HTTPPort)
	}
	return nil
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "valid config", cfg: Config{GRPCAddr: "host:50051", HTTPPort: 8080}, wantErr: false},
		{name: "empty grpc addr", cfg: Config{GRPCAddr: "", HTTPPort: 8080}, wantErr: true},
		{name: "invalid port zero", cfg: Config{GRPCAddr: "host:50051", HTTPPort: 0}, wantErr: true},
		{name: "invalid port too high", cfg: Config{GRPCAddr: "host:50051", HTTPPort: 99999}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
