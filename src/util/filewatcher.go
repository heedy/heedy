/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"

	log "github.com/Sirupsen/logrus"
)

// A Reloader has "reload" as a method
type Reloader interface {
	Reload() error
}

// FileWatcher watches a file for changes and calls "Reload" on the given interface. The FileWatcher
// also has an RWMutex for use in updating the file in a synchronous way
type FileWatcher struct {
	sync.RWMutex
	Reloader Reloader
	FileName string
	Watcher  *fsnotify.Watcher
	done     chan bool
}

// NewFileWatcher generates a new watcher for the given file path. It is assumed that the file was
// checked for existence already, and the Reloader is initialized with file contents already.
func NewFileWatcher(filename string, r Reloader) (*FileWatcher, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watch.Add(filename)
	if err != nil {
		watch.Close()
		return nil, err
	}

	done := make(chan bool)

	f := &FileWatcher{
		Reloader: r,
		FileName: filename,
		Watcher:  watch,
		done:     done,
	}

	go f.Watch()

	return f, nil

}

// Watch for changes and run reload when changes were detected.
func (f *FileWatcher) Watch() {
	for {
		select {
		case event := <-f.Watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Infof("Reloading '%s'", f.FileName)
				if err := f.Reloader.Reload(); err != nil {
					log.Warn(err.Error())
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				log.Warningf("File '%s' removed.", f.FileName)
				f.Watcher.Remove(f.FileName)

				// Keep trying to see if the file exists until it is found again
				for {
					time.Sleep(500 * time.Millisecond)
					v, err := os.Stat(f.FileName)
					if err == nil && !v.IsDir() {

						err = f.Watcher.Add(f.FileName)
						if err == nil {
							log.Infof("Reloading '%s'", f.FileName)
							err = f.Reloader.Reload()
							if err == nil {
								break
							}
							log.Warn(err.Error())
						}

					}
				}

			}
		case err := <-f.Watcher.Errors:
			log.Errorf("Watcher for '%s' failed: %s", f.FileName, err.Error())
			return
		case <-f.done:
			log.Debugf("Stopping file watch for '%s'", f.FileName)
			return
		}
	}
}

// Close shuts down the file watcher
func (f *FileWatcher) Close() {
	f.done <- true
	f.Watcher.Close()
}
