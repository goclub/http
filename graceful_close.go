package xhttp

import (
	"os"
	"os/signal"
	"syscall"
)

func GracefulClose(closeFunc func()) {
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-exit
	closeFunc()
}

