// Package redis provides Redis connection management.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"agent-comm-hub/internal/config"
)

// Manager manages Redis connections.
type Manager struct {
	standard *redis.Client
	pubsub   *redis.Client
}

// NewManager creates a new Redis manager.
func NewManager(cfg *config.RedisConfig) (*Manager, error) {
	standard, err := newRedisClient(cfg.StandardURL, cfg.PoolSize, cfg.MinIdleConn, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard Redis client: %w", err)
	}

	pubsub, err := newRedisClient(cfg.PubSubURL, cfg.PoolSize, cfg.MinIdleConn, cfg.Timeout)
	if err != nil {
		standard.Close()
		return nil, fmt.Errorf("failed to create pubsub Redis client: %w", err)
	}

	return &Manager{
		standard: standard,
		pubsub:   pubsub,
	}, nil
}

func newRedisClient(url string, poolSize, minIdleConn int, timeout time.Duration) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	opts.PoolSize = poolSize
	opts.MinIdleConns = minIdleConn
	opts.ReadTimeout = timeout
	opts.WriteTimeout = timeout
	opts.DialTimeout = timeout

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// Standard returns the standard Redis client.
func (m *Manager) Standard() *redis.Client {
	return m.standard
}

// PubSub returns the pub/sub Redis client.
func (m *Manager) PubSub() *redis.Client {
	return m.pubsub
}

// Close closes all Redis connections.
func (m *Manager) Close() error {
	var errs []error

	if err := m.standard.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := m.pubsub.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close Redis connections: %v", errs)
	}

	return nil
}

// Ping checks Redis connectivity.
func (m *Manager) Ping(ctx context.Context) error {
	if err := m.standard.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("standard Redis ping failed: %w", err)
	}
	if err := m.pubsub.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("pubsub Redis ping failed: %w", err)
	}
	return nil
}
