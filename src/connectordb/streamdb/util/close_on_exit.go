package util

import (
	"os"
	"os/signal"
	"syscall"
)

type Closeable interface {
	Close()
}

// CloseOnExit closes a resource when the program is exiting.
func CloseOnExit(closeable Closeable) {
	if closeable == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		closeable.Close()
	}()
}

// SendCloseSignal sends the program the terminate signal so all items waiting for a
// CloseOnExit will complete.
func SendCloseSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
