/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/kardianos/osext"
	"gopkg.in/fsnotify.v1"

	log "github.com/Sirupsen/logrus"
)

var (
	//The prefix to use for the paths in web server
	WWWPrefix = "www"
	AppPrefix = "app"

	// WWWPath is the path to the not-logged-in website files
	WWWPath string
	// AppPath is the path to the logged-in user website files
	AppPath string

	//These are the pre-loaded templates for non-logged in users
	WWWIndex *FileTemplate
	WWWLogin *FileTemplate
	WWWJoin  *FileTemplate
	WWW404   *FileTemplate

	//These are the pre-loaded templates for logged in users
	AppIndex  *FileTemplate
	AppUser   *FileTemplate
	AppDevice *FileTemplate
	AppStream *FileTemplate
	App404    *FileTemplate
	AppError  *FileTemplate
)

//FileTemplate implements all the necessary logic to read/write a "special" templated file
// as well as to update it from the folder in real time as it is modified.
type FileTemplate struct {
	sync.RWMutex //RWMutex allows for writing the template during runtime

	FilePath string
	Template *template.Template

	Watcher *fsnotify.Watcher
	done    chan bool
}

//NewFileTemplate loads a template from file and subscribes to changes from the file system
func NewFileTemplate(fpath string, err error) (*FileTemplate, error) {
	if err != nil {
		return nil, err
	}
	f, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Warnf("Could not read '%s'", fpath)
		return nil, err
	}
	tmpl, err := template.New(fpath).Parse(string(f))
	if err != nil {
		log.Warnf("Failed to parse '%s'", fpath)
		return nil, err
	}

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watch.Add(fpath)
	if err != nil {
		watch.Close()
	}

	done := make(chan bool)

	ft := &FileTemplate{
		RWMutex:  sync.RWMutex{},
		FilePath: fpath,
		Template: tmpl,
		Watcher:  watch,
		done:     done,
	}

	//Run the file watch in the background
	go ft.Watch()

	return ft, nil
}

// Reload loads up the template from the file path
func (f *FileTemplate) Reload() error {
	file, err := ioutil.ReadFile(f.FilePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New(f.FilePath).Parse(string(file))
	if err != nil {
		return fmt.Errorf("Failed to parse '%s': %v", f.FilePath, err.Error())
	}
	f.Lock()
	f.Template = tmpl
	f.Unlock()

	return nil
}

//Watch is run in the background to watch for changes in the template files
func (f *FileTemplate) Watch() {
	for {
		select {
		case event := <-f.Watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				//We reload the file
				log.Infof("Reloading file: '%s'", f.FilePath)
				err := f.Reload()
				if err != nil {
					log.Warn(err.Error())
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				log.Warningf("File '%s' removed. Using cached version.", f.FilePath)
				f.Watcher.Remove(f.FilePath)

				// Keep trying to see if the file exists until it is found again
				for {
					time.Sleep(2 * time.Second)
					v, err := os.Stat(f.FilePath)
					if err == nil && !v.IsDir() {

						err = f.Watcher.Add(f.FilePath)
						if err == nil {
							log.Infof("Reloading file: '%s'", f.FilePath)
							err := f.Reload()
							if err == nil {
								go f.Watch()
								return
							}
							log.Warn(err.Error())
						}

					}
				}

			}
		case err := <-f.Watcher.Errors:
			log.Errorf("Watcher for '%s' failed: %s", f.FilePath, err.Error())
			return
		case <-f.done:
			return
		}
	}
}

//Execute the template
func (f *FileTemplate) Execute(w io.Writer, data interface{}) error {
	f.RLock()
	err := f.Template.Execute(w, data)
	f.RUnlock()
	return err
}

// Close shuts down the file template
func (f *FileTemplate) Close() {
	f.Watcher.Close()
	f.done <- true
}

//LoadFiles sets up all the necessary files
func LoadFiles() error {
	//Now set up the app and www folder paths and make sure they exist
	exefolder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	WWWPath = path.Join(exefolder, WWWPrefix)
	log.Debugf("Hosting www from '%s'", WWWPath)

	AppPath = path.Join(exefolder, AppPrefix)
	log.Debugf("Hosting app from '%s'", AppPath)

	WWWIndex, err = NewFileTemplate(path.Join(WWWPath, "index.html"), err)
	WWWLogin, err = NewFileTemplate(path.Join(WWWPath, "login.html"), err)
	WWW404, err = NewFileTemplate(path.Join(WWWPath, "404.html"), err)
	WWWJoin, err = NewFileTemplate(path.Join(WWWPath, "join.html"), err)

	AppIndex, err = NewFileTemplate(path.Join(AppPath, "index.html"), err)
	AppUser, err = NewFileTemplate(path.Join(AppPath, "user.html"), err)
	AppDevice, err = NewFileTemplate(path.Join(AppPath, "device.html"), err)
	AppStream, err = NewFileTemplate(path.Join(AppPath, "stream.html"), err)
	App404, err = NewFileTemplate(path.Join(AppPath, "404.html"), err)
	AppError, err = NewFileTemplate(path.Join(AppPath, "error.html"), err)

	return err
}
