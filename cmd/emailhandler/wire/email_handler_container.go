package wire

import (
	"github.com/domesama/chat-and-notifications/connections"
	"github.com/domesama/chat-and-notifications/email"
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/domesama/doakes/doakeswire"
	doakes "github.com/domesama/doakes/server"
	"github.com/google/wire"
)

//go:generate wireprovider -source_root ../../../emailhandler/ -out ../wire/email_handler_providers.go -go_module_name "github.com/domesama/chat-and-notifications/emailhandler" -out_package "github.com/domesama/chat-and-notifications/cmd/emailhandler/wire"

type EmailHandlerContainer struct {
	*doakes.TelemetryServer
	httpserverwrapper.HTTPServer
}

func (r *EmailHandlerContainer) GetMonitoringServer() *doakes.TelemetryServer {
	return r.TelemetryServer
}

var MainBindingSet = wire.NewSet(
	LibSet,
	ConnectionSet,

	ProviderSet,
	httpserverwrapper.ProvideHTTPConfig,
	httpserverwrapper.ProvideHTTPServer,

	wire.Struct(new(EmailHandlerContainer), "*"),
)

var LibSet = wire.NewSet(
	doakeswire.TelemetrySetWithAutoStart,
)

var ConnectionSet = wire.NewSet(
	connections.MongoSet,
)

var RealSet = wire.NewSet(
	email.ProvideEmailConfig,
	email.ProvideSMTPEmailSender,
)
