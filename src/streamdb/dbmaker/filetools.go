package dbmaker

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kardianos/osext"
)

var (
	//ErrAlreadyRunning is thrown when a database that is already running is started
	ErrAlreadyRunning = errors.New("It looks like the database is already running. If you know it isn't, remove connectordb.pid")

	//ErrFileNotFound thrown when can't find a necessary file
	ErrFileNotFound = errors.New("A required configuration file was not found")

	//FolderPermissions is the folder permissions to use when creating a new database
	FolderPermissions = os.FileMode(0755)

	//FilePermissions refers to the permissions given to a file that is created
	FilePermissions = os.FileMode(0755)
)

//ConfigPath returns the path to the default StreamDB config templates
func ConfigPath(err error) (string, error) {
	if err != nil {
		return "", err
	}
	execpath, err := osext.ExecutableFolder()
	return filepath.Join(execpath, "config"), err
}

//PathExists returns whether or not the given path exists
func PathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return false
}

//IsDirectory returns if the path is a directory
func IsDirectory(filePath string) bool {
	s, err := os.Stat(filePath)
	if err == nil {
		return s.IsDir()
	}
	return false
}

//Touch creates a new file if it does not exist
func Touch(filePath string) error {
	if !PathExists(filePath) {
		return ioutil.WriteFile(filePath, []byte{}, FilePermissions)
	}
	return nil
}

//HasPidFile returns whether a PID file exists for the given process name
func HasPidFile(streamdbPath, processname string) bool {
	return PathExists(filepath.Join(streamdbPath, processname+".pid"))
}

//IsRunning returns whether connectordb is running for the given directory by checking for existence
//of a pid file
func IsRunning(streamdbPath string) bool {
	return HasPidFile(streamdbPath, "connectordb")
}

//EnsurePidNotRunning returns an error if it detects there is a pid file for connectordb already in the folder
func EnsurePidNotRunning(streamdbPath, processname string, err error) error {
	if err != nil {
		return err
	}
	if HasPidFile(streamdbPath, processname) {
		return ErrAlreadyRunning
	}
	return nil
}

//EnsureNotRunning returns an error if it detects there is a pid file for connectordb already in the folder
func EnsureNotRunning(streamdbPath string, err error) error {
	return EnsurePidNotRunning(streamdbPath, "connectordb", err)
}

//GetDatabaseType gets the database type used from the folder structure - in particular, if sqlite is used, then there
//will be an sqlite database. If a postgres folder exists, then dbtype is postgres. It returns ErrUnrecognizedDatabase
//if no database is recognized
func GetDatabaseType(streamdbDirectory string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if PathExists(filepath.Join(streamdbDirectory, sqliteDatabaseName)) {
		return "sqlite", nil
	}
	if PathExists(filepath.Join(streamdbDirectory, postgresDatabaseName)) {
		return "postgres", nil
	}
	return "", ErrUnrecognizedDatabase
}

//GenerateConfigReplacements generates the replacement variables to use within configuration files
func GenerateConfigReplacements(streamdbDirectory, procname, iface string, port int) map[string]string {
	m := make(map[string]string)

	if len(iface) == 0 {
		iface = "127.0.0.1"
	}

	m["dbdir"] = streamdbDirectory
	m["port"] = strconv.Itoa(port)
	m["interface"] = iface
	m["logfilepath"] = filepath.Join(streamdbDirectory, procname+".log")
	m["logfile"] = procname + ".log"
	m["pidfilepath"] = filepath.Join(streamdbDirectory, procname+".pid")
	m["pidfile"] = procname + ".pid"

	return m
}

//CopyConfig copies configuration file template from the default config directory of StreamDB to the database folder
func CopyConfig(streamdbPath, configname string, err error) error {
	if err != nil {
		return err
	}

	templatepath := filepath.Join(streamdbPath, configname)
	cpath, err := ConfigPath(err)
	defaultTemplate := filepath.Join(cpath, configname)
	if !PathExists(defaultTemplate) || err != nil {
		return ErrFileNotFound
	}
	log.Printf("Copying %s from '%s'", configname, defaultTemplate)
	return copyFileContents(defaultTemplate, templatepath, err)
}

//SetConfig sets up the given config file with the setting replacements. If a config template is
//not present in the streamdbpath, it looks in the root executable config directory for templates
func SetConfig(streamdbPath, configname string, replacements map[string]string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	log.Printf("Writing %s", configname)

	templatepath := filepath.Join(streamdbPath, configname)

	if !PathExists(templatepath) {
		err = CopyConfig(streamdbPath, configname, err)
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

	return outfile, err
}

//COPIED FROM: http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string, err error) error {
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}
