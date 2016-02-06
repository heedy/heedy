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
	"path"
	"util"

	"github.com/kardianos/osext"

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
	Template *template.Template

	Watcher *util.FileWatcher
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

	ft := &FileTemplate{
		Template: tmpl,
	}

	ft.Watcher, err = util.NewFileWatcher(fpath, ft)

	return ft, err
}

// Reload loads up the template from the file path
func (f *FileTemplate) Reload() error {
	file, err := ioutil.ReadFile(f.Watcher.FileName)
	if err != nil {
		return err
	}

	tmpl, err := template.New(f.Watcher.FileName).Parse(string(file))
	if err != nil {
		err = fmt.Errorf("Failed to parse '%s': %v", f.Watcher.FileName, err.Error())
		return err
	}
	f.Watcher.Lock()
	f.Template = tmpl
	f.Watcher.Unlock()

	return nil
}

//Execute the template
func (f *FileTemplate) Execute(w io.Writer, data interface{}) error {
	f.Watcher.RLock()
	err := f.Template.Execute(w, data)
	f.Watcher.RUnlock()
	return err
}

// Close shuts down the file template
func (f *FileTemplate) Close() {
	f.Watcher.Close()
}

//LoadFiles sets up all the necessary files
func LoadFiles() error {
	//Now set up the app and www folder paths and make sure they exist
	exefolder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	WWWPath = path.Join(exefolder, WWWPrefix)
	log.Infof("Hosting www from '%s'", WWWPath)

	AppPath = path.Join(exefolder, AppPrefix)
	log.Infof("Hosting app from '%s'", AppPath)

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
