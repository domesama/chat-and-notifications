package eventstore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/domesama/chat-and-notifications/event/eventmsg"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type (
	RedisEventStore[MsgValue any] struct {
		RedisClient redis.Client
		RedisEventStoreConfig
	}

	RedisEventStoreConfig struct {
		DeduplicationTTL    time.Duration
		EventStoreKeyPrefix string
	}
)

// ProvideRedisEventStore creates a new Redis-based event store for deduplication
// This is kept for convenience but can be constructed directly
func ProvideRedisEventStore(client redis.Client, conf RedisEventStoreConfig) EventStore[any] {
	return RedisEventStore[any]{
		RedisClient:           client,
		RedisEventStoreConfig: conf,
	}
}

func ProvideRedisEventStoreConfig() (conf RedisEventStoreConfig) {
	envconfig.MustProcess("", &conf)
	return
}

// FilterInvalidMessage checks if a message has already been processed
func (r RedisEventStore[MsgValue]) FilterInvalidMessage(
	ctx context.Context,
	msg eventmsg.Message[MsgValue],
) (filteredMessage eventmsg.Message[MsgValue], shouldDropEntirely bool) {
	// Create deduplication key from message key (should contain message_id:stream_id)
	dedupKey := fmt.Sprintf("eventstore:%s:%s", r.EventStoreKeyPrefix, msg.Key)

	exists, err := r.RedisClient.Exists(ctx, dedupKey).Result()
	if err != nil {
		slog.ErrorContext(
			ctx, "failed to check deduplication key in Redis",
			"error", err,
			"key", dedupKey,
		)
		// On error, allow processing to avoid blocking messages
		return msg, false
	}

	if exists > 0 {
		return msg, true
	}

	return msg, false
}

func (r RedisEventStore[MsgValue]) WriteEventStore(
	ctx context.Context,
	msg eventmsg.Message[MsgValue]) (err error) {
	dedupKey := fmt.Sprintf("eventstore:%s:%s", r.EventStoreKeyPrefix, msg.Key)

	err = r.RedisClient.Set(ctx, dedupKey, "1", r.DeduplicationTTL).Err()
	return
}
