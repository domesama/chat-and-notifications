package main

import (
	"log/slog"

	"github.com/domesama/chat-and-notifications/cmd/emailhandler/wire"
	"github.com/domesama/chat-and-notifications/utils"
)

func main() {
	ctn, cleanup, err := wire.StartEmailHandlerContainer()
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
