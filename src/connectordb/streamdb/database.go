package streamdb

import (
	"connectordb/config"
	"connectordb/streamdb/authoperator"
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/datastream/rediscache"
	"connectordb/streamdb/dbutil"
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
	"database/sql"
	"errors"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//The StreamDB version string
const (
	Version   = "0.2.2"
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
	operator.Operator //We need to do some magic so that the functions in Operator catch

	Userdb users.UserDatabase //SqlUserDatabase holds the methods needed to CRUD users/devices/streams

	ds  *datastream.DataStream //datastream holds methods for inserting datapoints into streams
	msg *Messenger             //messenger is a connection to the messaging client

	sqldb *sql.DB //We only need the sql object here to close it properly, since it is used everywhere.
}

/**
Open StreamDB given an Options object, which holds the information necessary to connect to the database

Finally, if running in postgres, then at least one process must be running the function RunWriter(). This function
writes the database's internal data.
**/
func Open(opt *config.Options) (dbp *Database, err error) {
	var db Database

	log.Debugln("Starting StreamDB")
	log.Debugln(opt.String())

	//Dbutil prints the sqluri to log, so no need to do it here
	db.sqldb, _, err = dbutil.OpenSqlDatabase(opt.SqlConnectionString)
	if err != nil {
		return nil, err
	}

	log.Debugln("Opening User database")
	db.Userdb = users.NewUserDatabase(db.sqldb, opt.SqlConnectionType, EnableCaching)

	log.Debugln("Opening messenger")
	db.msg, err = ConnectMessenger(&opt.NatsOptions, err)
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

	// Magic: Allows using the Database object as an operator.
	db.Operator = operator.Operator{&db}

	return &db, nil

}

//DeviceLoginOperator returns the operator associated with the given API key
func (db *Database) DeviceLoginOperator(devicepath, apikey string) (operator.Operator, error) {
	dev, err := db.ReadDevice(devicepath)
	if err != nil || dev.ApiKey != apikey {
		return operator.Operator{}, authoperator.ErrPermissions //Don't leak whether the device exists
	}
	return authoperator.NewAuthOperator(db, dev.DeviceId)
}

// UserLoginOperator returns the operator associated with the given username/password combination
func (db *Database) UserLoginOperator(username, password string) (operator.Operator, error) {
	usr, err := db.ReadUser(username)
	if err != nil || !usr.ValidatePassword(password) {
		return operator.Operator{}, authoperator.ErrPermissions //We don't want to leak if a user exists or not
	}

	dev, err := db.ReadDeviceByUserID(usr.UserId, "user")
	if err != nil {
		return operator.Operator{}, authoperator.ErrPermissions
	}

	return authoperator.NewAuthOperator(db, dev.DeviceId)
}

// LoginOperator logs in as a user or device, depending on which is passed in
func (db *Database) LoginOperator(path, password string) (operator.Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return operator.Operator{}, operator.ErrBadPath
	case 1:
		return db.DeviceLoginOperator(path, password)
	case 0:
		return db.UserLoginOperator(path, password)
	}
}

//Operator gets the operator by usr or device name
func (db *Database) GetOperator(path string) (operator.Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return operator.Operator{}, operator.ErrBadPath
	case 0:
		path += "/user"
	case 1:
		//Do nothing for this case
	}
	dev, err := db.ReadDevice(path)
	if err != nil {
		return operator.Operator{}, err //We use dev.Name, so must return error earlier
	}
	return authoperator.NewAuthOperator(db, dev.DeviceId)
}

//DeviceOperator returns the operator for the given device ID
func (db *Database) DeviceOperator(deviceID int64) (operator.Operator, error) {
	return authoperator.NewAuthOperator(db, deviceID)
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
	db.ds.RunWriter()
}

//Clear clears the database (to be used for debugging purposes - NEVER in production)
func (db *Database) Clear() {
	db.ds.Clear()
	db.sqldb.Exec("DELETE FROM Users;")
	db.sqldb.Exec("DELETE FROM Devices;")
	db.sqldb.Exec("DELETE FROM Streams;")
}

//These functions allow the Database object to conform to the BaseOperatorInterface

//Name here is a special one meaning that it is the database administration operator
// It is not a valid username
func (db *Database) Name() string {
	return AdminName
}

//Reload for a full database purges the entire cache
func (db *Database) Reload() error {
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
