package streamdb

import (
	"database/sql"
	"errors"
	"log"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/messenger"
	"streamdb/timebatchdb"
	"streamdb/users"
)

//The StreamDB version string
const Version = "0.2.0a"

var (
	//BatchSize is the batch size that StreamDB uses for its batching process. See Database.RunWriter()
	BatchSize = 250
)

//Database is a StreamDB database object which holds the methods
type Database struct {
	Userdb users.UserDatabase //UserDatabase holds the methods needed to CRUD users/devices/streams

	tdb *timebatchdb.Database //timebatchdb holds methods for inserting datapoints into streams
	msg *messenger.Messenger  //messenger is a connection to the messaging client

	sqldb *sql.DB //We only need the sql object here to close it properly, since it is used everywhere.
}

// Calls open from the arguments in the given configuration
func OpenFromConfig(cfg *config.Configuration) (*Database, error) {
	redis := cfg.GetRedisUri()
	gnatsd := cfg.GetGnatsdUri()
	sql := cfg.GetDatabaseConnectionString()

	return Open(sql, redis, gnatsd)
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
	var db Database
	var sqltype string

	//Dbutil prints the sqluri t log, so no need to do it here
	db.sqldb, sqltype, err = dbutil.OpenSqlDatabase(sqluri)
	if err != nil {
		return nil, err
	}

	log.Printf("Opening User database")
	db.Userdb.InitUserDatabase(db.sqldb, sqltype)

	log.Printf("Opening messenger with uri %s\n", msguri)
	db.msg, err = messenger.Connect(msguri, err)

	log.Printf("Opening timebatchdb with redis url %v batch size: %v\n", redisuri, BatchSize)
	db.tdb, err = timebatchdb.Open(db.sqldb, sqltype, redisuri, BatchSize, err)

	if err != nil {
		db.Close()
		return nil, err
	}

	// If it is an sqlite database, run the timebatchdb writer (since it is guaranteed to be only process)
	if sqltype == "sqlite3" {
		go db.tdb.WriteDatabase()
	}

	return &db, nil

}

//DeviceOperator returns the operator associated with the given API key
func (db *Database) DeviceOperator(apikey string) (Operator, error) {
	dev, err := db.Userdb.ReadDeviceByApiKey(apikey)
	if err != nil {
		return nil, err
	}
	usr, err := db.Userdb.ReadUserById(dev.UserId)
	return &AuthOperator{db, dev, usr}, err
}

//UserOperator returns the operator associated with the given username/password combination
func (db *Database) UserOperator(username, password string) (Operator, error) {
	usr, dev, err := db.Userdb.Login(username, password)
	return &AuthOperator{db, dev, usr}, err
}

func (db *Database) Operator(path string) (Operator, error) {
	return nil, errors.New("UNIMPLEMENTED")
}

//Close closes all database connections and releases all resources.
//A word of warning though: If RunWriter() is functional, then RunWriter will crash
func (db *Database) Close() {
	if db.tdb != nil {
		db.tdb.Close()
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
	db.tdb.WriteDatabase()
}
