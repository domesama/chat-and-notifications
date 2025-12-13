package wire

import (
	"github.com/domesama/chat-and-notifications/connections"
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/domesama/doakes/doakeswire"
	doakes "github.com/domesama/doakes/server"
	"github.com/google/wire"
)

//go:generate wireprovider -source_root ../../../chatpersistence/ -out ../wire/chat_persistence_providers.go -go_module_name "github.com/domesama/chat-and-notifications/chatpersistence" -out_package "github.com/domesama/chat-and-notifications/cmd/chatpersistence/wire"

type ChatPersistenceContainer struct {
	*doakes.TelemetryServer
	httpserverwrapper.HTTPServer
}

func (r *ChatPersistenceContainer) GetMonitoringServer() *doakes.TelemetryServer {
	return r.TelemetryServer
}

var MainBindingSet = wire.NewSet(
	LibSet,
	ConnectionSet,

	ProviderSet,
	httpserverwrapper.ProvideHTTPConfig,
	httpserverwrapper.ProvideHTTPServer,

	wire.Struct(new(ChatPersistenceContainer), "*"),
)

var LibSet = wire.NewSet(
	doakeswire.TelemetrySetWithAutoStart,
)

var ConnectionSet = wire.NewSet(
	connections.MongoSet,
)
