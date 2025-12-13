package wire

import (
	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler"
	"github.com/domesama/doakes/doakeswire"
	doakes "github.com/domesama/doakes/server"
	"github.com/google/wire"
)

//go:generate wireprovider -source_root ../../../chatpersistencechangehandler/ -out ../wire/chat_persistence_change_handler_providers.go -go_module_name "github.com/domesama/chat-and-notifications/chatpersistencechangehandler" -out_package "github.com/domesama/chat-and-notifications/cmd/chatpersistencechangehandler/wire"

type ChatPersistenceChangeHandlerContainer struct {
	*doakes.TelemetryServer
	chatpersistencechangehandler.ChatPersistenceChangeHandler
}

func (r *ChatPersistenceChangeHandlerContainer) GetMonitoringServer() *doakes.TelemetryServer {
	return r.TelemetryServer
}

var MainBindingSet = wire.NewSet(
	LibSet,
	ProviderSet,

	wire.Struct(new(ChatPersistenceChangeHandlerContainer), "*"),
)

var LibSet = wire.NewSet(
	doakeswire.TelemetrySetWithAutoStart,
)
