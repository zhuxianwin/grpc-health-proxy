package health

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Checker performs a single gRPC health check against a target address.
type Checker interface {
	Check(ctx context.Context, service string) Result
}

// GRPCChecker dials a gRPC endpoint and calls the Health/Check RPC.
type GRPCChecker struct {
	addr string
}

// NewChecker returns a GRPCChecker that connects to addr.
func NewChecker(addr string) *GRPCChecker {
	return &GRPCChecker{addr: addr}
}

// Check performs a gRPC Health/Check call and maps the response to a Result.
func (c *GRPCChecker) Check(ctx context.Context, service string) Result {
	conn, err := grpc.DialContext(
		ctx,
		c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return Result{
			Service: service,
			Status:  StatusUnknown,
			Err:     fmt.Errorf("dial %s: %w", c.addr, err),
		}
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		return Result{
			Service: service,
			Status:  StatusUnknown,
			Err:     fmt.Errorf("health check rpc: %w", err),
		}
	}

	switch resp.GetStatus() {
	case grpc_health_v1.HealthCheckResponse_SERVING:
		return Result{Service: service, Status: StatusHealthy}
	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		return Result{Service: service, Status: StatusUnhealthy}
	default:
		return Result{
			Service: service,
			Status:  StatusUnknown,
			Err:     fmt.Errorf("unexpected status: %s", resp.GetStatus()),
		}
	}
}
