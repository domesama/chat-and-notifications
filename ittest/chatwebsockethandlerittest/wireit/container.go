package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/chatwebsocketshandler/wire"
	"github.com/google/wire"
)

type ChatWebSocketHandlerITTestContainer struct {
	applicationwire.Locator
	applicationwire.ChatWebSocketHandlerContainer
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,
	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(ChatWebSocketHandlerITTestContainer), "*"),
)
