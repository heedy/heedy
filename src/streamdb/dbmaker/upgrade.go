package dbmaker

import (
	"errors"
)

//Upgrade the database
func Upgrade(streamdbDirectory string, err error) error {
	return errors.New("Upgrade is not defined for your database version")
}
