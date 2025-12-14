package chatpersistencechangehandler

import (
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/domesama/chat-and-notifications/eventstore"
	"github.com/redis/go-redis/v9"
)

type ChatPersistenceChangeEventStore struct {
	eventstore.RedisEventStore[eventmodel.ChatMessagePersistenceChangeEvent]
}

func ProvideChatPersistenceChangeEventStore(
	client redis.Client,
	conf eventstore.RedisEventStoreConfig) (store ChatPersistenceChangeEventStore) {

	store.RedisClient = client
	store.RedisEventStoreConfig = conf

	return

}
