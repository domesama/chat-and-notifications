package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/chatpersistence/wire"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatPersistenceITTestContainer struct {
	applicationwire.Locator
	applicationwire.ChatPersistenceContainer
	*mongo.Database
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,
	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(ChatPersistenceITTestContainer), "*"),
)
