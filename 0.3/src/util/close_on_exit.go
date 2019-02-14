/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

var (
	closeMutex  = sync.Mutex{}
	closeExiter = sync.Once{}
	closers     = []Closeable{}
	closeWaiter = sync.WaitGroup{}
)

//CloseCall calls a custom function on close
type CloseCall struct {
	Callme func()
}

//Close calls the function
func (c CloseCall) Close() {
	c.Callme()
}

//Closeable is anything that can be closed
type Closeable interface {
	Close()
}

// CloseOnExit closes a resource when the program is exiting.
func CloseOnExit(closeable Closeable) {
	closeExiter.Do(setupCloseOnExit)

	closeMutex.Lock()
	closers = append(closers, closeable)
	closeMutex.Unlock()
}

// SetupCloseOnExit sets up the close on exit code
func setupCloseOnExit() {
	c := make(chan os.Signal, 3)

	if runtime.GOOS != "windows" {
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	} else {
		signal.Notify(c, os.Interrupt)
	}

	go func() {
		for {
			s := <-c
			switch s {
			case syscall.SIGTERM, os.Interrupt:
				log.Warn("Exiting...")
				closeMutex.Lock()
				closeWaiter.Add(len(closers))
				log.Debugf("Running %d cleanup tasks", len(closers))
				for i := range closers {
					curnum := i

					go func() {
						defer func() {
							if r := recover(); r != nil {
								log.Errorf("Cleanup task panic: %s", debug.Stack())
							}
						}()
						closers[curnum].Close()
						log.Debugf("Done with cleanup task #%d", curnum+1)
						closeWaiter.Done()
					}()

				}
				closeMutex.Unlock()
				closeWaiter.Wait()
				log.Warn("bye!")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Warn("Caught SIGHUP - ignoring")
			}
		}

	}()
}
