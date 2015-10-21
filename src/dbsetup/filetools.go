package dbsetup

import (
	"config"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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

	m["dbdir"] = serviceDirectory
	m["port"] = strconv.Itoa(int(s.Port))
	m["interface"] = s.Hostname
	m["logfilepath"] = filepath.Join(serviceDirectory, procname+".log")
	m["logfile"] = procname + ".log"
	m["pidfilepath"] = filepath.Join(serviceDirectory, procname+".pid")
	m["pidfile"] = procname + ".pid"

	m["username"] = s.Username
	m["password"] = s.Password

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
