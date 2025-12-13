package wire

import (
	"github.com/domesama/chat-and-notifications/connections"
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/domesama/doakes/doakeswire"
	doakes "github.com/domesama/doakes/server"
	"github.com/google/wire"
)

//go:generate wireprovider -source_root ../../../chatwebsocketshandler/ -out ../wire/chat_websocket_handler_providers.go -go_module_name "github.com/domesama/chat-and-notifications/chatwebsocketshandler" -out_package "github.com/domesama/chat-and-notifications/cmd/chatwebsocketshandler/wire"

type ChatWebSocketHandlerContainer struct {
	httpserverwrapper.HTTPWithWebSocketServer
	*doakes.TelemetryServer
}

func (r *ChatWebSocketHandlerContainer) GetMonitoringServer() *doakes.TelemetryServer {
	return r.TelemetryServer
}

var MainBindingSet = wire.NewSet(
	LibSet,
	ConnectionSet,

	ProviderSet,

	websocket.ProvideWebSocketConfig,
	websocket.ProvideDefaultWebSocketManager,

	httpserverwrapper.ProvideHTTPConfig,
	httpserverwrapper.ProvideHTTPWithWebSocketServer,

	wire.Struct(new(ChatWebSocketHandlerContainer), "*"),
)

var LibSet = wire.NewSet(
	doakeswire.TelemetrySetWithAutoStart,
)

var ConnectionSet = wire.NewSet(
	connections.MongoSet,
)
