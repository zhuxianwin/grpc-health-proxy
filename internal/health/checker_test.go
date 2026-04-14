package health

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func startFakeHealthServer(t *testing.T, serving bool) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	hsrv := health.NewServer()
	status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if serving {
		status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	hsrv.SetServingStatus("", status)
	grpc_health_v1.RegisterHealthServer(s, hsrv)

	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	return lis.Addr().String()
}

func TestChecker_Healthy(t *testing.T) {
	addr := startFakeHealthServer(t, true)
	checker := NewChecker(addr, 2*time.Second, 0)

	ok, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected healthy=true")
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	addr := startFakeHealthServer(t, false)
	checker := NewChecker(addr, 2*time.Second, 0)

	ok, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected healthy=false")
	}
}

func TestChecker_Unreachable(t *testing.T) {
	checker := NewChecker("127.0.0.1:1", 200*time.Millisecond, 0)

	_, err := checker.Check(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestChecker_CacheHit(t *testing.T) {
	addr := startFakeHealthServer(t, true)
	checker := NewChecker(addr, 2*time.Second, 10*time.Second)

	ok1, err := checker.Check(context.Background(), "")
	if err != nil || !ok1 {
		t.Fatalf("first check failed: err=%v healthy=%v", err, ok1)
	}

	// Second call should return cached result without hitting the server.
	ok2, err := checker.Check(context.Background(), "")
	if err != nil || !ok2 {
		t.Fatalf("cached check failed: err=%v healthy=%v", err, ok2)
	}
}
