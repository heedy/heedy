/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package util

import (
	"os"
	"path/filepath"
	"os/exec"
	"errors"
	"io"
	"io/ioutil"
	"time"
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

	//This error is thrown if a bad path is given as a database
	ErrNotDatabase = errors.New("The given path is not initialized as a database")
)


// Sets the present working directory to the path of the executable
func SetWdToExecutable() error {
	path, _ := exec.LookPath(os.Args[0])
	fp, _ := filepath.Abs(path)
	dir, _ := filepath.Split(fp)
	return os.Chdir(dir)
}



// IsDirectory returns if the path is a directory, or false on error
func IsDirectory(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return s.IsDir()
}


/**
	Checks that the given directory exists, is a connectordb directory,
	and isn't currently running; then returns the absolute path to it.
**/
func ProcessDatabaseDirectory(directory string) (string, error) {
	if IsDirectory(directory) {
		directory, _ = filepath.Abs(directory)
	} else {
		return directory, ErrNotDatabase
	}

	err := EnsureNotRunning(directory)
	return directory, err
}



//PathExists returns whether or not the given path exists
func PathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return false
}


//Touch creates a new file if it does not exist
func Touch(filePath string) error {
	if !PathExists(filePath) {
		return ioutil.WriteFile(filePath, []byte{}, FilePermissions)
	}

	// unix touch also updates the file times.
	return os.Chtimes(filePath, time.Now(), time.Now())
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
func EnsureNotRunning(streamdbPath string) error {
	return EnsurePidNotRunning(streamdbPath, "connectordb", nil)
}



//COPIED FROM: http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContents(src, dst string, err error) error {
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
