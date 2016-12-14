/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"util"

	log "github.com/Sirupsen/logrus"
	"github.com/kardianos/osext"
)

var (

	//ErrFileNotFound thrown when can't find a necessary file
	ErrFileNotFound = errors.New("A required configuration file was not found")

	//FolderPermissions is the folder permissions to use when creating a new database
	FolderPermissions = os.FileMode(0755)

	//FilePermissions refers to the permissions given to a file that is created
	FilePermissions = os.FileMode(0755)
)

var funcMap = template.FuncMap{
	"add": func(a, b int64) int64 {
		return a + b
	},
	"mul": func(a, b int64) int64 {
		return a * b
	},
}

//GenerateConfigReplacements generates the replacement variables to use within configuration files
func GenerateConfigReplacements(serviceDirectory, procname string, c *config.Configuration) map[string]interface{} {
	m := make(map[string]interface{})

	serviceDirectory2, err := filepath.Abs(serviceDirectory)
	if err == nil {
		serviceDirectory = serviceDirectory2
	}

	// The _slash variants are because sometimes the config files require unix style slashes even on windows
	// This means that we use the slash variant in those files.

	m["dbdir"] = serviceDirectory
	m["procname"] = procname
	m["dbdir_slash"] = filepath.ToSlash(serviceDirectory)

	lfp := filepath.Join(serviceDirectory, procname+".log")
	m["logfilepath"] = lfp
	m["logfilepath_slash"] = filepath.ToSlash(lfp)
	m["logfile"] = procname + ".log"
	pidf := filepath.Join(serviceDirectory, procname+".pid")
	m["pidfilepath"] = pidf
	m["pidfilepath_slash"] = filepath.ToSlash(pidf)
	m["pidfile"] = procname + ".pid"

	m["cdb"] = c

	return m
}

//ConfigPath returns the path to the default ConnectorDB config templates
func ConfigPath() (string, error) {
	execpath, err := osext.ExecutableFolder()
	return filepath.Join(execpath, "config"), err
}

//CopyConfig copies configuration file template from the default config directory of ConnectorDB to the database folder
func CopyConfig(servicePath, configname string, err error) error {
	if err != nil {
		return err
	}

	templatepath := filepath.Join(servicePath, configname)
	cpath, err := ConfigPath()
	defaultTemplate := filepath.Join(cpath, configname)
	if !util.PathExists(defaultTemplate) || err != nil {
		log.Errorf("Error path: %s configname: %s default: %s err: %v", servicePath, configname, defaultTemplate, err)
		return ErrFileNotFound
	}
	log.Debugf("Copying %s from '%s'", configname, defaultTemplate)
	return util.CopyFileContents(defaultTemplate, templatepath, err)
}

//SetConfig sets up the given config file template. If the configuration file is
//not present in the servicePath, it looks in the root executable config directory for templates
func SetConfig(servicePath, configname string, replacements map[string]interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}

	log.Debugf("Writing %s", configname)

	templatepath := filepath.Join(servicePath, configname)

	if !util.PathExists(templatepath) {
		err = CopyConfig(servicePath, configname, err)
	}

	configfilecontents, err := ioutil.ReadFile(templatepath)
	if err != nil {
		return "", err
	}

	t, err := template.New(templatepath).Funcs(funcMap).Parse(string(configfilecontents))
	if err != nil {
		return "", err
	}

	// Open the file to write
	outfile := templatepath + ".tmp"
	f, err := os.Create(outfile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	t.Execute(f, replacements)

	log.Debugf("wrote config file %s", outfile)
	return outfile, err
}

// GetPostgresExecutablePath is a hack to allow windows postgres to work. The installation
// must be set up with the correct directory structure, meaning that the postgres database
// will be in the pgsql directory. So we first check this directory!
func GetPostgresExecutablePath(executableName string) string {
	execpath, err := osext.ExecutableFolder()
	if err != nil {
		panic(err.Error())
	}

	//Check if the binaries are given in our executable folder
	execpath = filepath.Join(execpath, "dep", "pgsql", "bin", executableName)

	if runtime.GOOS == "windows" {
		execpath += ".exe"
	}
	log.Debugf("Checking for '%s'", execpath)
	if util.PathExists(execpath) {
		return execpath
	}
	p, err := getExecutablePath(executableName)
	if err == nil && p != "" {
		return p
	}

	if runtime.GOOS == "windows" {
		// NOPE - grep doesn't work in windows
		panic(fmt.Sprintf("Could not find executable %s", executableName))
	}

	// On ubuntu, postgres seems not to be in path by default. We therefore use grep-find
	p = findPostgresExecutableGrep(executableName)
	if p == "" {
		panic(fmt.Sprintf("Could not find executable %s", executableName))
	}
	return trimExecutablePath(p)
}

// Find a postgres utility e.g. initdb or postgres using the lame grep method, works on Ubuntu (for now)
func findPostgresExecutableGrep(executableName string) string {
	log.Debugf("Checking for %s by grep in /usr/lib/postgresql", executableName)

	findCmd := fmt.Sprintf("find /usr/lib/postgresql/ | sort -r | grep -m 1 /bin/%v", executableName)

	cmd := exec.Command("bash", "-c", findCmd)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return ""
	}

	return string(out)
}

// GetExecutablePath gets the path for the executable. This is a general version of the functions used for getting postgres
// executable
func GetExecutablePath(executableName string) string {
	s, err := getExecutablePath(executableName)
	if err != nil {
		panic(err.Error())
	}
	if s == "" {
		panic(fmt.Sprintf("Could not find executable %s", executableName))
	}
	return s
}

// getExecutablePath returns a string and an error instead of panic on failure to find
func getExecutablePath(executableName string) (string, error) {
	// A version of the executable in the dep folder takes prescedence over everything else
	execpath, err := osext.ExecutableFolder()
	if err != nil {
		return "", err
	}

	//Check if the binaries are given in our executable folder
	execpath = filepath.Join(execpath, "dep", executableName)

	if runtime.GOOS == "windows" {
		execpath += ".exe"
	}
	log.Debugf("Checking for '%s'", execpath)

	if util.PathExists(execpath) {
		return execpath, nil
	} else if runtime.GOOS == "windows" {
		return "", fmt.Errorf("Could not find executable %s", executableName)
	}
	log.Debugf("Checking for %s in path...", executableName)
	// Start with which because we prefer a PATH version
	out := trimExecutablePath(findExecutableWhich(executableName))
	if out != "" {
		log.Debugf("Using %s", out)
		return out, nil
	}

	return "", fmt.Errorf("Could not find executable %s", executableName)
}

// Finds a utility on $PATH
func findExecutableWhich(executableName string) string {
	cmd := exec.Command("which", executableName)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return ""
	}

	return string(out)
}

func trimExecutablePath(exepath string) string {
	return strings.Trim(exepath, " \t\n\r")
}
