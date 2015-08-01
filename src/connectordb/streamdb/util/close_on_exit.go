package util

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

var (
	//The number of goroutines waiting for a close signal
	closeNumber  = int32(0)
	closeCounter = make(chan bool, 1)
	closeExiter  = sync.Once{}
)

type Closeable interface {
	Close()
}

// CloseOnExit closes a resource when the program is exiting.
func CloseOnExit(closeable Closeable) {
	atomic.AddInt32(&closeNumber, 1)
	closeExiter.Do(func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-c
			log.Warn("Exiting...")
			cnum := atomic.LoadInt32(&closeNumber)
			for cnum > 0 {
				<-closeCounter
				cnum = atomic.LoadInt32(&closeNumber)
			}
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
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt32(&closeNumber, -1)
				closeCounter <- false
			}
		}()
		<-c
		closeable.Close()
		atomic.AddInt32(&closeNumber, -1)
		closeCounter <- true
	}()
}

// SendCloseSignal sends the program the terminate signal so all items waiting for a
// CloseOnExit will complete.
func SendCloseSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
