package main

import (
	"log/slog"

	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/domesama/chat-and-notifications/cmd/chatpersistencechangehandler/wire"
	"github.com/domesama/chat-and-notifications/utils"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	var appConfig config.ChatPersistenceChangeHandlerConfig
	envconfig.MustProcess("", &appConfig)

	ctn, cleanup, err := wire.StartChatPersistenceChangeHandlerContainer(appConfig)
	defer cleanup()
	if err != nil {
		slog.Error("cannot initialize server")
		panic(err)
	}

	if ctn.GetMonitoringServer() != nil {
		ctn.GetMonitoringServer().EnableHealthCheck()
	}

	utils.WaitForTerminatingSignal()
}
