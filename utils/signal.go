package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForTerminatingSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
