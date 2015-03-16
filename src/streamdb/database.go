package streamdb

import (
	"database/sql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"log"
	"streamdb/dtypes"
	"streamdb/users"
	"strings"
)

const (
	SQLITE_PREFIX   = "sqlite://"
	POSTGRES_PREFIX = "postgres://"
)

var (
	BATCH_SIZE = 100 //The batch size that StreamDB uses for its batching process. See Database.RunWriter()
)

//This is a StreamDB database object which holds the methods
type Database struct {
	users.UserDatabase                       //UserDatabase holds the methods needed to CRUD users/devices/streams
	tdb                *dtypes.TypedDatabase //timebatchdb holds methods for inserting datapoints into streams

	sqldb *sql.DB //Connection to the sql database
}

//This function closes all database connections and releases all resources.
//A word of warning though: If RunWriter() is functional, then RunWriter will crash
func (db *Database) Close() {
	if db.tdb != nil {
		db.tdb.Close()
	}

	if db.Db != nil {
		db.Db.Close()
	}
}

/**
Opens StreamDB given urls to the SQL database used, to the redis instance and to the gnatsd messenger
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

	//First, we check if the user wants to use sqlite or postgres. If the url given
	//has the hallmarks of a file or sqlite database, then set that as the database type

	db.SqlType = users.POSTGRES //The default is postgres.

	switch {
	// TODO just check if this is a file
	// How I wish there were a "blah" in [] like Python!
	case strings.HasSuffix(sqluri, ".db") ||
		strings.HasSuffix(sqluri, ".sqlite") ||
		strings.HasSuffix(sqluri, ".sqlite3") ||
		strings.HasPrefix(sqluri, SQLITE_PREFIX):

		db.SqlType = users.SQLITE3

		//The sqlite driver doesn't like starting with sqlite://
		if strings.HasPrefix(sqluri, SQLITE_PREFIX) {
			sqluri = sqluri[len(SQLITE_PREFIX):]
		}
		break
	case strings.HasPrefix(sqluri, POSTGRES_PREFIX):
		db.SqlType = users.POSTGRES
		sqluri = sqluri[len(POSTGRES_PREFIX):]
	}

	/*TODO: Right now UserDB has no way to pass in an sql object without
	  bypassing all constructors.
	  So we let UserDB open the connection, then steal the database object
	  //Now open the correct type of database.
	  sdb,err := sql.Open(sqltype,sqluri)
	  if err!=nil {
	      return nil,err
	  }
	*/

	log.Printf("Opening %v database with cxn string: %v", db.SqlType, sqluri)
	err = db.InitUserDatabase(db.SqlType, sqluri)
	db.sqldb = db.Db

	log.Printf("Opening timebatchdb with redis url %v batch size: %v", redisuri, BATCH_SIZE)
	db.tdb, err = dtypes.Open(db.Db, string(db.SqlType), redisuri, BATCH_SIZE, err)

	if err != nil {
		db.Close()
		return nil, err
	}

	// If it is an sqlite database, run the timebatchdb writer (since it is guaranteed to be only process)
	if db.SqlType == users.SQLITE3 {
		go db.tdb.WriteDatabase()
	}

	return &db, nil

}

/**
StreamDB uses a batching mechanism for writing timestamps, where data is first written to redis, and then committed to
an sql database in batches of size BATCH_SIZE (global var). This allows great insert speed as well as fantastic read speed on large
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
