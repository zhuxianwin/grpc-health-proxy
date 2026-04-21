package server_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/example/grpc-health-proxy/internal/config"
	"github.com/example/grpc-health-proxy/internal/server"
)

func freePort() int { return 19876 }

// startServer creates and starts a server with the given config, returning a
// cleanup function that shuts the server down. It fails the test immediately
// if the server cannot be created.
func startServer(t *testing.T, cfg *config.Config) *server.Server {
	t.Helper()
	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	go func() { _ = srv.Start() }()
	time.Sleep(50 * time.Millisecond)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	return srv
}

func TestServer_LivezEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort:    freePort(),
		HTTPPath:    "/healthz",
		GRPCAddress: "localhost:50051",
		GRPCService: "",
	}

	startServer(t, cfg)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/livez", cfg.HTTPPort))
	if err != nil {
		t.Fatalf("GET /livez error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.Config{
		HTTPPort:    freePort() + 1,
		HTTPPath:    "/healthz",
		GRPCAddress: "localhost:50051",
		GRPCService: "",
	}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	go func() { _ = srv.Start() }()
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}
