package wire

import (
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/domesama/doakes/doakeswire"
	doakes "github.com/domesama/doakes/server"
	"github.com/google/wire"
)

//go:generate wireprovider -source_root ../../../generalnotifications/ -out ../wire/general_notification_handler_providers.go -go_module_name "github.com/domesama/chat-and-notifications/generalnotifications" -out_package "github.com/domesama/chat-and-notifications/cmd/generalnotificationshandler/wire"

type GeneralNotificationHandlerContainer struct {
	httpserverwrapper.HTTPWithWebSocketServer
	*doakes.TelemetryServer
}

func (r *GeneralNotificationHandlerContainer) GetMonitoringServer() *doakes.TelemetryServer {
	return r.TelemetryServer
}

var MainBindingSet = wire.NewSet(
	LibSet,

	ProviderSet,

	websocket.ProvideWebSocketConfig,
	websocket.ProvideDefaultWebSocketManager,

	httpserverwrapper.ProvideHTTPConfig,
	httpserverwrapper.ProvideHTTPWithWebSocketServer,

	wire.Struct(new(GeneralNotificationHandlerContainer), "*"),
)

var LibSet = wire.NewSet(
	doakeswire.TelemetrySetWithAutoStart,
)
