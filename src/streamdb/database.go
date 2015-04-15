package streamdb

import (
	"database/sql"
	"log"
	"streamdb/messenger"
	"streamdb/timebatchdb"
	"streamdb/users"
	"streamdb/dbutil"
)



var (
	//BatchSize is the batch size that StreamDB uses for its batching process. See Database.RunWriter()
	BatchSize = 250
	db *Database
)

//Database is a StreamDB database object which holds the methods
type Database struct {
	users.UserDatabase //UserDatabase holds the methods needed to CRUD users/devices/streams

	tdb   *timebatchdb.Database //timebatchdb holds methods for inserting datapoints into streams
	msg   *messenger.Messenger  //messenger is a connectino to the messaging client
	sqldb *sql.DB               //Connection to the sql database
	SqlType dbutil.DRIVERSTR

	dbutil.SqlxMixin
}


/**
Open StreamDB given urls to the SQL database used, to the redis instance and to the gnatsd messenger
server.

StreamDB can use both postgres and sqlite as its storage engine. To run StreamDB with sqlite, give a
path to the database file ending with .db. If the file does not end in .db, use "sqlite://" in the path
to make sure that is the database engine used:
  streamdb.Open("sqlite://path/to/database","localhost:6379","localhost:4222")
One important thing to note when running StreamDB with sqlite: Open() automatically starts RunWriter() in a goroutine
on open, since it is assumed that this is the only object from which the database is accessed.

The normal use case for StreamDB is postgres. For postgres, just use the url of the connection. If you are worried
that StreamDB will mistake your url for a sqlite location, you can start your database string with "postgres://".
An example of a postgres url will be:
  streamdb.Open("postgres://username:password@localhost:port/databasename?sslmode=verify-full","localhost:6379","localhost:4222")
If just running locally, then you can use:
  streamdb.Open("postgres://localhost:52592/connectordb?sslmode=disable","localhost:6379","localhost:4222")
The preceding command will use the database "connectordb" (which is assumed to have been created already) on the local machine.

Finally, if running in postgres, then at least one process must be running the function RunWriter(). This function
writes the database's internal data.
**/
func Open(sqluri, redisuri, msguri string) (dbp *Database, err error) {

	/**
	  TODO migrate all sql userdb stuff into this file.
	  **/

	var db Database


	db.sqldb, db.SqlType, err = dbutil.OpenSqlDatabase(sqluri)

	if err != nil {
		db.Close()
		return nil, err
	}

	db.InitSqlxMixin(db.sqldb, string(db.SqlType))
	db.InitUserDatabase(db.sqldb, string(db.SqlType))

	log.Printf("Opening messenger with uri %s\n", msguri)
	db.msg, err = messenger.Connect(msguri, err)

	log.Printf("Opening timebatchdb with redis url %v batch size: %v\n", redisuri, BatchSize)
	db.tdb, err = timebatchdb.Open(db.sqldb, string(db.SqlType), redisuri, BatchSize, err)

	if err != nil {
		db.Close()
		return nil, err
	}

	// If it is an sqlite database, run the timebatchdb writer (since it is guaranteed to be only process)
	if db.SqlType == dbutil.SQLITE3 {
		go db.tdb.WriteDatabase()
	}

	return &db, nil

}


//Close closes all database connections and releases all resources.
//A word of warning though: If RunWriter() is functional, then RunWriter will crash
func (db *Database) Close() {
	if db.tdb != nil {
		db.tdb.Close()
	}

	if db.sqldb != nil {
		db.sqldb.Close()
	}

	if db.msg != nil {
		db.msg.Close()
	}
}

/**
RunWriter exists because StreamDB uses a batching mechanism for writing timestamps, where data is first written to redis, and then committed to
an sql database in batches of size BatchSize (global var). This allows great insert speed as well as fantastic read speed on large
ranges of data. RunWriter runs this 'batching' process, which happens in the background.
When running a single instance with posgres, you need to call RunWriter once manually (as a goroutine). You do not need to
run it if on sqlite, since it is run automatically.
If running as a cluster, then it is probably a good idea to have RunWriter be run as an entirely separate process.

For example:
  db,_ := streamdb.Open("postgres://...",...)
  go db.RunWriter()   //Run this right after starting StreamDB
  ...
  db.Close()

If unsure as to whether you should call RunWriter, this is a good way to decide:
Are you running StreamDB manually by yourself using postgres, and this is the only process? If so then yes.
If you are just connecting to an already-running StreamDB and RunWriter is already running somewhere on
this database, then NO.
**/
func (db *Database) RunWriter() {
	if db.SqlType != "sqlite3" {
		db.tdb.WriteDatabase()
	}
}
