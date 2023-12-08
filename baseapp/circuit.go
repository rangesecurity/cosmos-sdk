package baseapp

import (
	"context"
	"time"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, blockTime time.Time, typeURL string) (bool, error)
}
