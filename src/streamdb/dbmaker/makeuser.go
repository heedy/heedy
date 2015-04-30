package dbmaker

import (
	"log"
	"streamdb/dbutil"
	"streamdb/users"
	"streamdb/util"
)

//MakeUser creates a user directly from a streamdb directory, without needing to start up all of streamdb
func MakeUser(streamdbDirectory, username, password, email string, err error) error {
	if err != nil {
		return err
	}

	streamdbDirectory, err = util.ProcessConnectordbDirectory(streamdbDirectory)
	if err != nil {
		return err
	}

	//Start the postgres database on a random port on localhost to set up the user
	err = StartSqlDatabase(streamdbDirectory, "127.0.0.1", 55413, err)

	if err != nil {
		return err
	}

	log.Printf("Creating user %s (%s)\n", username, email)

	spath, err := GetSqlPath(streamdbDirectory, "127.0.0.1", 55413, err)

	db, driver, err := dbutil.OpenSqlDatabase(spath)
	if err == nil {
		var udb users.UserDatabase
		udb.InitUserDatabase(db, string(driver))
		err = udb.CreateUser(username, email, password)
	}
	StopSqlDatabase(streamdbDirectory, nil)
	return err
}
