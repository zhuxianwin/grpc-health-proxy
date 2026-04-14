package health

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Status represents the health check result.
type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "HEALTHY"
	case StatusUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// Checker performs gRPC health checks against a target service.
type Checker struct {
	addr    string
	timeout time.Duration
}

// NewChecker creates a new Checker for the given address.
func NewChecker(addr string, timeout time.Duration) *Checker {
	return &Checker{
		addr:    addr,
		timeout: timeout,
	}
}

// Check performs a gRPC health check for the given service name.
// Pass an empty string to check the overall server health.
func (c *Checker) Check(ctx context.Context, service string) (Status, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return StatusUnhealthy, fmt.Errorf("failed to connect to %s: %w", c.addr, err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: service,
	})
	if err != nil {
		return StatusUnhealthy, fmt.Errorf("health check rpc failed: %w", err)
	}

	if resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
		return StatusHealthy, nil
	}
	return StatusUnhealthy, nil
}
