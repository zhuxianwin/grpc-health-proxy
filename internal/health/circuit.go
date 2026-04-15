package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing fast
	CircuitHalfOpen                     // testing recovery
)

// CircuitConfig holds configuration for the circuit breaker.
type CircuitConfig struct {
	FailureThreshold int
	SuccessThreshold int
	OpenDuration     time.Duration
}

// DefaultCircuitConfig returns sensible defaults.
func DefaultCircuitConfig() CircuitConfig {
	return CircuitConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenDuration:     30 * time.Second,
	}
}

// circuitBreaker wraps a Checker with circuit breaker logic.
type circuitBreaker struct {
	inner   Checker
	cfg     CircuitConfig
	mu      sync.Mutex
	state   CircuitState
	failures int
	successes int
	openedAt time.Time
}

// NewCircuitBreaker wraps inner with a circuit breaker.
func NewCircuitBreaker(inner Checker, cfg CircuitConfig) Checker {
	return &circuitBreaker{inner: inner, cfg: cfg}
}

func (c *circuitBreaker) Check(ctx context.Context, service string) Result {
	c.mu.Lock()
	state := c.resolveState()
	c.mu.Unlock()

	if state == CircuitOpen {
		return Result{Status: StatusUnhealthy, Err: fmt.Errorf("circuit open for service %q", service)}
	}

	res := c.inner.Check(ctx, service)

	c.mu.Lock()
	defer c.mu.Unlock()

	if res.Err != nil || res.Status == StatusUnhealthy {
		c.failures++
		c.successes = 0
		if c.state == CircuitHalfOpen || c.failures >= c.cfg.FailureThreshold {
			c.state = CircuitOpen
			c.openedAt = time.Now()
		}
	} else {
		c.successes++
		c.failures = 0
		if c.state == CircuitHalfOpen && c.successes >= c.cfg.SuccessThreshold {
			c.state = CircuitClosed
		}
	}
	return res
}

// resolveState transitions Open -> HalfOpen after OpenDuration; must hold mu.
func (c *circuitBreaker) resolveState() CircuitState {
	if c.state == CircuitOpen && time.Since(c.openedAt) >= c.cfg.OpenDuration {
		c.state = CircuitHalfOpen
		c.successes = 0
	}
	return c.state
}
