package health_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	checkerPkg "github.com/yourorg/grpc-health-proxy/internal/health"
)

func startFakeHealthServer(t *testing.T, serving bool) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	healthSrv := health.NewServer()

	status := grpc_health_v1.HealthCheckResponse_SERVING
	if !serving {
		status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	healthSrv.SetServingStatus("", status)
	grpc_health_v1.RegisterHealthServer(server, healthSrv)

	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(server.GracefulStop)

	return lis.Addr().String()
}

func TestChecker_Healthy(t *testing.T) {
	addr := startFakeHealthServer(t, true)
	checker := checkerPkg.NewChecker(addr, 3*time.Second)

	status, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != checkerPkg.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", status)
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	addr := startFakeHealthServer(t, false)
	checker := checkerPkg.NewChecker(addr, 3*time.Second)

	status, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != checkerPkg.StatusUnhealthy {
		t.Errorf("expected UNHEALTHY, got %s", status)
	}
}

func TestChecker_Unreachable(t *testing.T) {
	checker := checkerPkg.NewChecker("127.0.0.1:1", 500*time.Millisecond)

	status, err := checker.Check(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for unreachable address, got nil")
	}
	if status != checkerPkg.StatusUnhealthy {
		t.Errorf("expected UNHEALTHY, got %s", status)
	}
}
