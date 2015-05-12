package streamdb

import (
	"database/sql"
	"errors"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/messenger"
	"streamdb/timebatchdb"
	"streamdb/users"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//The StreamDB version string
const (
	Version   = "0.2.0"
	AdminName = "  Admin  "
)

var (

	//ErrBadPath is thrown when a user device or stream cannot be extracted from the path
	ErrBadPath = errors.New("The given path is invalid.")
	//ErrAdmin is thrown when trying to get the user or device of the Admin operator
	ErrAdmin = errors.New("An administrative operator has no user or device")

	//BatchSize is the batch size that StreamDB uses for its batching process. See Database.RunWriter()
	BatchSize = 250

	//CacheSizes are the number of users/devices/streams to cache
	UserCacheSize   = 100
	DeviceCacheSize = 1000
	StreamCacheSize = 10000

	//Cache eviction timer in seconds. Can be huge on single-node setups
	CacheExpireTime = 60
)

//Database is a StreamDB database object which holds the methods
type Database struct {
	Userdb users.UserDatabase //UserDatabase holds the methods needed to CRUD users/devices/streams

	tdb *timebatchdb.Database //timebatchdb holds methods for inserting datapoints into streams
	msg *messenger.Messenger  //messenger is a connection to the messaging client

	sqldb *sql.DB //We only need the sql object here to close it properly, since it is used everywhere.

	//The caches are to keep frequently used stuff in memory for a reasonable time before reloading from database
	userCache   *TimedCache
	deviceCache *TimedCache
	streamCache *TimedCache
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

	//Dbutil prints the sqluri to log, so no need to do it here
	db.sqldb, sqltype, err = dbutil.OpenSqlDatabase(sqluri)
	if err != nil {
		return nil, err
	}

	log.Debugln("Opening User database")
	db.Userdb.InitUserDatabase(db.sqldb, sqltype)

	log.Debugln("Opening messenger with uri ", msguri)
	db.msg, err = messenger.Connect(msguri, err)

	log.Debugf("Opening timebatchdb with redis url %v batch size: %v", redisuri, BatchSize)
	db.tdb, err = timebatchdb.Open(db.sqldb, sqltype, redisuri, BatchSize, err)

	db.userCache, err = NewTimedCache(UserCacheSize, int64(CacheExpireTime), err)
	db.deviceCache, err = NewTimedCache(DeviceCacheSize, int64(CacheExpireTime), err)
	db.streamCache, err = NewTimedCache(StreamCacheSize, int64(CacheExpireTime), err)

	if err != nil {
		db.Close()
		return nil, err
	}

	// If it is an sqlite database, run the timebatchdb writer (since it is guaranteed to be only process)
	if sqltype == config.Sqlite {
		go db.tdb.WriteDatabase()
	}

	return &db, nil

}

//DeviceLoginOperator returns the operator associated with the given API key
func (db *Database) DeviceLoginOperator(devicepath, apikey string) (Operator, error) {
	dev, err := db.ReadDevice(devicepath)
	if err != nil || dev.ApiKey != apikey {
		return nil, ErrPermissions //Don't leak whether the device exists
	}

	return &AuthOperator{db, devicepath, dev.DeviceId}, err
}

//UserLoginOperator returns the operator associated with the given username/password combination
func (db *Database) UserLoginOperator(username, password string) (Operator, error) {
	usr, err := db.ReadUser(username)
	if err != nil || !usr.ValidatePassword(password) {
		return nil, ErrPermissions //We don't want to leak if a user exists or not
	}

	dev, err := db.ReadDeviceByUserID(usr.UserId, "user")
	if err != nil {
		return nil, ErrPermissions //We use dev.Name, so must return error earlier
	}

	return &AuthOperator{db, usr.Name + "/" + dev.Name, dev.DeviceId}, nil
}

//LoginOperator logs in as a user or device, depending on which is passed in
func (db *Database) LoginOperator(path, password string) (Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return nil, ErrBadPath
	case 1:
		return db.DeviceLoginOperator(path, password)
	case 0:
		return db.UserLoginOperator(path, password)
	}
}

//Operator gets the operator by usr or device name
func (db *Database) Operator(path string) (Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return nil, ErrBadPath
	case 0:
		path += "/user"
	case 1:
		//Do nothing for this case
	}
	dev, err := db.ReadDevice(path)
	if err != nil {
		return nil, err //We use dev.Name, so must return error earlier
	}
	return &AuthOperator{db, path, dev.DeviceId}, err
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

//These functions allow the Database object to conform to the Operator interface

//Name here is a special one meaning that it is the database administration operator
func (db *Database) Name() string {
	return AdminName
}

//Database just returns self
func (db *Database) Database() *Database {
	return db
}

//Reload for a full database purges the entire cache
func (db *Database) Reload() error {
	db.userCache.Purge()
	db.deviceCache.Purge()
	db.streamCache.Purge()
	return nil
}

//User returns the current user
func (db *Database) User() (usr *users.User, err error) {
	return nil, ErrAdmin
}

//Device returns the current device
func (db *Database) Device() (*users.Device, error) {
	return nil, ErrAdmin
}

//Permissions returns whether the operator has permissions given by the string
func (db *Database) Permissions(perm users.PermissionLevel) bool {
	return true
}

//SetAdmin does exactly what it claims. It works on both users and devices
func (db *Database) SetAdmin(path string, isadmin bool) error {
	switch strings.Count(path, "/") {
	default:
		return ErrBadPath
	case 0:
		u, err := db.ReadUser(path)
		if err != nil {
			return err
		}
		u.Admin = isadmin
		return db.UpdateUser(u)
	case 1:
		dev, err := db.ReadDevice(path)
		if err != nil {
			return err
		}
		dev.IsAdmin = isadmin
		return db.UpdateDevice(dev)
	}
}
