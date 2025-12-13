package connections

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
	"github.com/redis/go-redis/v9"
)

// ProvideRedisClient creates a Redis client with the provided configuration.
// Returns the client, a cleanup function, and an error.
func ProvideRedisClient(cfg connectionconfig.RedisClientConfig) (redis.Client, func(), error) {
	client := redis.NewClient(
		&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		},
	)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return redis.Client{}, func() {}, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	slog.Info(
		"Redis client connected",
		"addr", cfg.Addr,
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	cleanup := func() {
		if err := client.Close(); err != nil {
			slog.Error("failed to close Redis client", "error", err)
		} else {
			slog.Info("Redis client closed")
		}
	}

	return *client, cleanup, nil
}
