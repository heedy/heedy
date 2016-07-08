/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

//GenerateConfigReplacements generates the replacement variables to use within configuration files
func GenerateConfigReplacements(serviceDirectory, procname string, s *config.Service) map[string]string {
	m := make(map[string]string)

	serviceDirectory2, err := filepath.Abs(serviceDirectory)
	if err == nil {
		serviceDirectory = serviceDirectory2
	}

	if len(s.Hostname) == 0 {
		s.Hostname = "localhost"
	}

	// The _slash variants are because sometimes the config files require unix style slashes even on windows
	// This means that we use the slash variant in those files.

	m["dbdir"] = serviceDirectory
	m["dbdir_slash"] = filepath.ToSlash(m["dbdir"])
	m["port"] = strconv.Itoa(int(s.Port))
	m["interface"] = s.Hostname
	m["logfilepath"] = filepath.Join(serviceDirectory, procname+".log")
	m["logfilepath_slash"] = filepath.ToSlash(m["logfilepath"])
	m["logfile"] = procname + ".log"
	m["pidfilepath"] = filepath.Join(serviceDirectory, procname+".pid")
	m["pidfilepath_slash"] = filepath.ToSlash(m["pidfilepath"])
	m["pidfile"] = procname + ".pid"

	m["username"] = s.Username
	m["password"] = s.Password

	// Some config files might require linux-style path names

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

//SetConfig sets up the given config file with the setting replacements. If a config template is
//not present in the servicePath, it looks in the root executable config directory for templates
func SetConfig(servicePath, configname string, replacements map[string]string, err error) (string, error) {
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

	//Replace stuff in the config file
	for key, value := range replacements {
		configfilecontents = []byte(strings.Replace(string(configfilecontents), "{{"+key+"}}", value, -1))
	}

	outfile := templatepath + ".tmp"
	err = ioutil.WriteFile(outfile, configfilecontents, FilePermissions)

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
	return GetExecutablePath(executableName)
}

// GetExecutablePath gets the path for the executable. This is a general version of the functions used for getting postgres
// executable
func GetExecutablePath(executableName string) string {
	// A version of the executable in the dep folder takes prescedence over everything else
	execpath, err := osext.ExecutableFolder()
	if err != nil {
		panic(err.Error())
	}

	//Check if the binaries are given in our executable folder
	execpath = filepath.Join(execpath, "dep", executableName)

	if runtime.GOOS == "windows" {
		execpath += ".exe"
	}
	log.Debugf("Checking for '%s'", execpath)

	if util.PathExists(execpath) {
		return execpath
	} else if runtime.GOOS == "windows" {
		panic(fmt.Sprintf("Could not find executable %s", executableName))
	}
	log.Debugf("Checking for %s in path...", executableName)
	// Start with which because we prefer a PATH version
	out := findExecutableWhich(executableName)

	if out != "" {
		log.Debugf("Using %s", out)
		return trimExecutablePath(out)
	}

	panic(fmt.Sprintf("Could not find executable %s", executableName))
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
