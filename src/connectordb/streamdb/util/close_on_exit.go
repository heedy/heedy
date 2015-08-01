package util

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

var (
	closeWaiter = sync.WaitGroup{}
	closeExiter = sync.Once{}
)

type Closeable interface {
	Close()
}

// CloseOnExit closes a resource when the program is exiting.
func CloseOnExit(closeable Closeable) {
	closeWaiter.Add(1)
	closeExiter.Do(func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-c
			log.Warn("Exiting...")
			closeWaiter.Wait()
			log.Debug("bye!")
			os.Exit(0)
		}()
	})

	if closeable == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		closeable.Close()
		closeWaiter.Done()
	}()
}

// SendCloseSignal sends the program the terminate signal so all items waiting for a
// CloseOnExit will complete.
func SendCloseSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
