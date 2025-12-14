package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/chatpersistencechangehandler/wire"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

type ChatPersistenceChangeHandlerITTestContainer struct {
	applicationwire.ChatPersistenceChangeHandlerContainer
	applicationwire.Locator

	RedisClient redis.Client
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,
	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(ChatPersistenceChangeHandlerITTestContainer), "*"),
)
