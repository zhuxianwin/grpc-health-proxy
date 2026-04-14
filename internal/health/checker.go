package health

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_healthRPC health checks against service.
type Checker struct string
	cache   *Cache
	tout time.Duration
}
 Checker targeting the given gRPC address.
// cacheTTL controls how long results are cached; use 0 to disable caching.
func NewChecker(addr string, timeout, cacheTTL time.Duration) *Checker {
	return &Checker{
		addr:    addr,
		cache:   NewCache(cacheTTL),
		timeout: timeout,
	}
}

// Check queries the gRPC Health service for the given service name.
// An empty service name checks the overall server health.
func (c *Checker) Check(ctx context.Context, service string) (bool, error) {
	if cached, ok := c.cache.Get(service); ok {
		return cached.Healthy, nil
	}

	conn, err := grpc.NewClient(
		c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false, fmt.Errorf("dial %s: %w", c.addr, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: service,
	})
	if err != nil {
		c.cache.Set(service, false)
		return false, fmt.Errorf("health check rpc: %w", err)
	}

	healthy := resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING
	c.cache.Set(service, healthy)
	return healthy, nil
}
