/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package connectordb

import (
	"config"
	"connectordb/datastream"
	"connectordb/datastream/rediscache"
	"connectordb/operator/messenger"
	"connectordb/users"
	"database/sql"
	"dbsetup/dbutil"
	"errors"
	"testing"
	"util"

	log "github.com/Sirupsen/logrus"
)

//The StreamDB version string
const (
	Version   = "0.3.0a"
	AdminName = " ADMIN "
)

var (
	//ErrAdmin is thrown when trying to get the user or device of the Admin operator
	ErrAdmin = errors.New("An administrative operator has no user or device")

	//BatchSize is the batch size that StreamDB uses for its batching process. See Database.RunWriter()
	BatchSize = 100

	EnableCaching = true
)

//Database is a StreamDB database object which holds the methods
type Database struct {
	Userdb users.UserDatabase //SqlUserDatabase holds the methods needed to CRUD users/devices/streams

	ds  *datastream.DataStream //datastream holds methods for inserting datapoints into streams
	msg *messenger.Messenger   //messenger is a connection to the messaging client

	sqldb *sql.DB //We only need the sql object here to close it properly, since it is used everywhere.
}

// These are here for Operator to be able to construct properly, they cannot
// have pointer receivers.
func (db Database) GetUserDatabase() users.UserDatabase {
	return db.Userdb
}
func (db Database) GetDatastream() *datastream.DataStream {
	return db.ds
}
func (db Database) GetMessenger() *messenger.Messenger {
	return db.msg
}

/**
Open StreamDB given an Options object, which holds the information necessary to connect to the database

Finally, if running in postgres, then at least one process must be running the function RunWriter(). This function
writes the database's internal data.
**/
func Open(opt *config.Options) (dbp *Database, err error) {
	var db Database

	log.Debugln("Opening ConnectorDB")
	log.Debugln(opt.String())

	//Dbutil prints the sqluri to log, so no need to do it here
	db.sqldb, _, err = dbutil.OpenSqlDatabase(opt.SqlConnectionString)
	if err != nil {
		return nil, err
	}

	log.Debugln("Opening User database")
	db.Userdb = users.NewUserDatabase(db.sqldb, config.SqlType, EnableCaching)

	log.Debugln("Opening messenger")
	db.msg, err = messenger.ConnectMessenger(&opt.NatsOptions, err)
	if err != nil {
		return nil, err
	}

	log.Debugf("Opening Redis cache")
	rc, err := rediscache.NewRedisConnection(&opt.RedisOptions)
	if err != nil {
		db.Close()
		return nil, err
	}
	rc.BatchSize = int64(opt.BatchSize)

	log.Debugf("Opening DataStream")
	db.ds, err = datastream.OpenDataStream(rediscache.RedisCache{rc}, db.sqldb, opt.ChunkSize)
	if err != nil {
		rc.Close()
		db.Close()
		return nil, err
	}

	// Close the database when the system exits just in case it isn't.
	util.CloseOnExit(&db)

	return &db, nil

}

//Close closes all database connections and releases all resources.
//A word of warning though: If RunWriter() is functional, then RunWriter will crash
func (db *Database) Close() {
	if db.ds != nil {
		db.ds.Close()
	}
	if db.msg != nil {
		db.msg.Close()
	}
	if db.sqldb != nil {
		db.sqldb.Close()
	}
}

/**
RunWriter exists because StreamDB uses a batching mechanism for writing timestamps, where data is first written to redis, and then committed to
an sql database in batches of size BatchSize (global var). This allows great insert speed as well as fantastic read speed on large
ranges of data. RunWriter runs this 'batching' process, which happens in the background.
When running a single instance with posgres, you need to call RunWriter once manually (as a goroutine).
If running as a cluster, then it is probably a good idea to have RunWriter be run as an entirely separate process.

For example:
  db,_ := connectordb.Open("postgres://...",...)
  go db.RunWriter()   //Run this right after starting StreamDB
  ...
  db.Close()

If unsure as to whether you should call RunWriter, this is a good way to decide:
Are you running StreamDB manually by yourself using postgres, and this is the only process? If so then yes.
If you are just connecting to an already-running StreamDB and RunWriter is already running somewhere on
this database, then NO.
**/
func (db *Database) RunWriter() error {
	return db.ds.RunWriter()
}

// Clear clears the database (to be used for debugging purposes - NEVER in production)
// the reason it's in this file!
func (db *Database) Clear(t testing.TB) {
	if t == nil {
		return
	}

	db.ds.Clear()
	db.sqldb.Exec("DELETE FROM Users;")
	db.sqldb.Exec("DELETE FROM Devices;")
	db.sqldb.Exec("DELETE FROM Streams;")
}
