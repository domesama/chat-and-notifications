package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/chatpersistencechangehandler/wire"
	"github.com/google/wire"
)

type ChatPersistenceChangeHandlerITTestContainer struct {
	applicationwire.ChatPersistenceChangeHandlerContainer
	applicationwire.Locator
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,
	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(ChatPersistenceChangeHandlerITTestContainer), "*"),
)
