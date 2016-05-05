/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"connectordb"
	"html/template"
	"io"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/gernest/hot"
	"github.com/kardianos/osext"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

const (
	//The prefix to use for the paths in web server
	WWWPrefix = "www"
	AppPrefix = "app"
)

type wwwtemplatebookmark string
type apptemplatebookmark string

var (

	// WWWPath is the path to the not-logged-in website files
	WWWPath string
	// AppPath is the path to the logged-in user website files
	AppPath string

	// These are the pre-loaded templates for non-logged in users
	WWWTemplate *hot.Template
	// These are the pre-loaded templates for logged in users
	AppTemplate *hot.Template

	// These are convenience functions for accessing specific endpoints
	WWWLogin  wwwtemplatebookmark = "login.html"
	WWWIndex  wwwtemplatebookmark = "index.html"
	WWW404    wwwtemplatebookmark = "404.html"
	WWWJoin   wwwtemplatebookmark = "join.html"
	AppIndex  apptemplatebookmark = "index.html"
	AppUser   apptemplatebookmark = "user.html"
	AppDevice apptemplatebookmark = "device.html"
	AppStream apptemplatebookmark = "stream.html"
	AppError  apptemplatebookmark = "error.html"
)

func (w wwwtemplatebookmark) Execute(wr io.Writer, data interface{}) (err error) {
	return WWWTemplate.Execute(wr, string(w), data)
}

func (w apptemplatebookmark) Execute(wr io.Writer, data interface{}) (err error) {
	return AppTemplate.Execute(wr, string(w), data)
}

func isBlank(text string) bool {
	return text == ""
}

func dataURIToAttr(uri string) template.HTMLAttr {
	return template.HTMLAttr("src=\"" + uri + "\"")
}

func markdown(input string) template.HTML {
	unsafe := blackfriday.MarkdownCommon([]byte(input))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(string(html))
}

func version() string {
	return connectordb.Version
}

// defaultMarkdown returns HTML formatted first text if it is not blank,
// otherwise it returns markdown formatted second text
func defaultMarkdown(first, defaultText string) template.HTML {
	if !isBlank(first) {
		return markdown(first)
	}
	return markdown(defaultText)
}

//LoadFiles sets up all the necessary files
func LoadFiles() error {
	funcMap := template.FuncMap{
		"isblank":      isBlank,
		"dataURIToSrc": dataURIToAttr,
		"Version":      version,
		"markdown":     markdown,
		"default":      defaultMarkdown,
	}

	//Now set up the app and www folder paths and make sure they exist
	exefolder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	WWWPath = path.Join(exefolder, WWWPrefix)
	log.Infof("Hosting www from '%s'", WWWPath)

	AppPath = path.Join(exefolder, AppPrefix)
	log.Infof("Hosting app from '%s'", AppPath)

	{
		config := &hot.Config{
			Watch:          true,
			BaseName:       "index",
			Dir:            WWWPath,
			FilesExtension: []string{".html"},
			Funcs:          funcMap,
		}

		WWWTemplate, err = hot.New(config)
		if err != nil {
			return err
		}
	}

	{
		config := &hot.Config{
			Watch:          true,
			BaseName:       "index",
			Dir:            AppPath,
			FilesExtension: []string{".html"},
			Funcs:          funcMap,
		}

		AppTemplate, err = hot.New(config)
		if err != nil {
			return err
		}
	}

	return err
}
