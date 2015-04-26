package dbmaker

import (
	"errors"
	"path/filepath"
)

//Upgrade the database
func Upgrade(streamdbDirectory string, err error) error {
	if err == nil {
		if IsDirectory(streamdbDirectory) {
			streamdbDirectory, err = filepath.Abs(streamdbDirectory)
		} else {
			return ErrNotDatabase
		}

	}

	err = EnsureNotRunning(streamdbDirectory, err)
	if err != nil {
		return err
	}

	return errors.New("Upgrade is not defined for your database version")
}
