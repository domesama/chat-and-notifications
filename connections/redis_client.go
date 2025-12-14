package connections

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

var RedisSet = wire.NewSet(
	connectionconfig.ProvideRedisClientConfig, ProvideRedisClient,
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

	cleanup := func() {
		if err := client.Close(); err != nil {
			slog.Error("failed to close Redis client", "error", err)
		}
	}

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return redis.Client{}, cleanup, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return *client, cleanup, nil
}
