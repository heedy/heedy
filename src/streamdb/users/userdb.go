// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	_ "github.com/lib/pq"
	"log"
)


const(
	// A black and qhite question mark
	DEFAULT_ICON =`iVBORw0KGgoAAAANSUhEUgAAAEAAAABAAQMAAACQp+OdAAAABlBMVEUAA
	AAAAAClZ7nPAAAAAXRSTlMAQObYZgAAAAFiS0dEAIgFHUgAAAAJcEhZcwAACxMAAAsTAQCanBgAAACVS
	URBVCjPjdGxDcQgDAVQRxSUHoFRMtoxWkZhBEoKK/+IsaNc0ElQxE8K3xhBtLa4Gj4YNQBFEYHxjwFRJ
	OBU7AAsZOgVWSEJR68bajSUoOjfoK07NkP+h/jAiI8g2WgGdqRx+jVa/r0P2cx9EPE2zduUVxv2NHs6n
	Q6Z0BZQaX3F4/0od3xvE2TCtOeOs12UQl6c5Quj42jQ5zt8GQAAAABJRU5ErkJggg==`
	DEFAULT_PASSWORD_HASH = "SHA512"
)

var (
	// Standard Errors
	ERR_EMAIL_EXISTS = errors.New("A user already exists with this email")
	ERR_USERNAME_EXISTS = errors.New("A user already exists with this username")
	ERR_INVALID_PTR = errors.New("The provided pointer is nil")


	// statements
	CREATE_PHONE_CARRIER_STMT = `INSERT INTO PhoneCarrier (Name, EmailDomain) VALUES (?,?);`
	SELECT_PHONE_CARRIER_BY_ID_STMT = "SELECT * FROM PhoneCarrier WHERE Id = ? LIMIT 1"
	UPDATE_PHONE_CARRIER_STMT = `UPDATE PhoneCarrier SET Name=?, EmailDomain=? WHERE Id = ?;`
	DELETE_PHONE_CARRIER_BY_ID_STMT = `DELETE FROM PhoneCarrier WHERE Id = ?;`
	CREATE_DEVICE_STMT = `INSERT INTO Device
	    (	Name,
	        ApiKey,
	        Icon_PngB64,
	        OwnerId)
	        VALUES (?,?,?,?)`

	SELECT_DEVICE_BY_USER_ID_STMT = "SELECT * FROM Device WHERE OwnerId = ?"
	SELECT_DEVICE_BY_ID_STMT = "SELECT * FROM Device WHERE Id = ? LIMIT 1"
	SELECT_DEVICE_BY_API_KEY_STMT = "SELECT * FROM Device WHERE ApiKey = ? LIMIT 1"
	UPDATE_DEVICE_STMT = `UPDATE Device SET
	    Name = ?, ApiKey = ?, Enabled = ?,
	    Icon_PngB64 = ?, Shortname = ?, Superdevice = ?,
	    OwnerId = ?, CanWrite = ?, CanWriteAnywhere = ?, UserProxy = ? WHERE Id = ?;`
	DELETE_DEVICE_BY_ID_STMT = `DELETE FROM Device WHERE Id = ?;`
	CREATE_STREAM_STMT = `INSERT INTO Stream
	    (	Name,
	        Type,
	        OwnerId) VALUES (?,?,?);`
	SELECT_STREAM_BY_ID_STMT = "SELECT * FROM Stream WHERE Id = ? LIMIT 1"
	SELECT_STREAM_BY_DEVICE_STMT = "SELECT * FROM Stream WHERE OwnerId = ?"
	UPDATE_STREAM_STMT = `UPDATE Stream SET
	    Name = ?,
	    Active = ?,
	    Public = ?,
	    Type = ?,
	    OwnerId = ?,
	    Ephemeral = ?,
	    Output = ?
	    WHERE Id = ?;`
	DELETE_STREAM_BY_ID_STMT = `DELETE FROM Stream WHERE Id = ?;`
	CREATE_USER_STMT = `INSERT INTO Users (
	    Name,
	    Email,
	    Password,
	    PasswordSalt,
	    PasswordHashScheme,
	    CreateTime) VALUES (?,?,?,?,?,?);`
	SELECT_USER_BY_EMAIL_STMT = "SELECT * FROM Users WHERE Email = ? LIMIT 1"
	SELECT_USER_BY_NAME_STMT = "SELECT * FROM Users WHERE Name = ? LIMIT 1"
	SELECT_USER_BY_ID_STMT = "SELECT * FROM Users WHERE Id = ? LIMIT 1"
	SELECT_ALL_USERS_STMT = "SELECT * FROM Users"
	SELECT_OWNER_OF_STREAM_BY_ID_STMT = `SELECT u.*
	                              FROM Users u, Stream s, Device d
	                              WHERE s.Id = ?
	                                AND d.Id = s.OwnerId
	                                AND u.Id = d.OwnerId
	                              LIMIT 1;`
	UPDATE_USER_STMT = `UPDATE Users SET
	                Name=?, Email=?, Password=?, PasswordSalt=?, PasswordHashScheme=?,
	                Admin=?, Phone=?, PhoneCarrier=?, UploadLimit_Items=?,
	                ProcessingLimit_S=?, StorageLimit_Gb=?, CreateTime = ?, ModifyTime = ?,
	                UserGroup = ? WHERE Id = ?;`
	DELETE_USER_BY_ID_STMT = `DELETE FROM Users WHERE Id = ?;`
	READ_DEVICE_BY_USER_AND_NAME = "SELECT * FROM Device WHERE OwnerId = ? AND Name = ?"
	READ_STREAM_BY_DEVICE_AND_NAME = "SELECT * FROM Stream WHERE OwnerId = ? AND Name = ?"
)

type DRIVERSTR string;

const(
	SQLITE3 DRIVERSTR = "sqlite3"
	POSTGRES DRIVERSTR = "postgres"
)

// The main UserDatabase type
type UserDatabase struct {
	SqlType DRIVERSTR
	filepath string
	port int
	//db *sql.DB // TODO remove this.
	Db *sql.DB
}


func NewSqliteUserDatabase(path string) (*UserDatabase, error) {
	n := new(UserDatabase)
	err := n.InitUserDatabase(SQLITE3, path)

	return n, err
}

func NewPostgresUserDatabase(cxnString string) (*UserDatabase, error) {
	n := new(UserDatabase)
	err := n.InitUserDatabase(POSTGRES, cxnString)
	return n, err
}


func (userdb *UserDatabase) InitUserDatabase(ds DRIVERSTR, cxn string) error {
	userdb.SqlType = ds
	userdb.filepath = cxn

	var err error
	userdb.Db, err = sql.Open(string(userdb.SqlType), userdb.filepath)
	//userdb.db = userdb.Db
	if err != nil {
		return err
	}

	err = userdb.Db.Ping()
	if err != nil {
		return err
	}

	switch ds {
		case SQLITE3:
			if err := userdb.setupSqliteDatabase(); err != nil {
				return err
			}
		case POSTGRES:
			if err := userdb.setupPostgresDatabase(); err != nil {
				return err
			}
		default:
			return errors.New("Illegal driver string")
	}

	return nil
}

func (userdb *UserDatabase) UnderlyingDb() *sql.DB {
	return userdb.Db
}




/**
Sets up the SQLITE databse.
**/
func (userdb *UserDatabase) setupSqliteDatabase() error{

	log.Printf("setting up sqlite db")

	_, err := userdb.Db.Exec("PRAGMA foreign_keys = ON;")

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier (
	    id integer primary key,
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);`);

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, _ = userdb.Db.Exec(`INSERT INTO PhoneCarrier VALUES (0, 'None', '')`);

    /** mysql
    CREATE TABLE IF NOT EXISTS User(   Id SERIAL PRIMARY KEY AUTO_INCREMENT, Name VARCHAR(50) UNIQUE NOT NULL, Email VARCHAR(100) UNIQUE NOT NULL, Password VARCHAR(100) NOT NULL, PasswordSalt VARCHAR(100) NOT NULL, PasswordHashScheme VARCHAR(50) NOT NULL, Admin BOOLEAN DEFAULT FALSE, Phone VARCHAR(50) DEFAULT "", PhoneCarrier INTEGER DEFAULT 0, UploadLimit_Items INTEGER DEFAULT 24000, ProcessingLimit_S INTEGER DEFAULT 86400, StorageLimit_Gb INTEGER DEFAULT 4,  FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL );
    **/
	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Users (Id INTEGER PRIMARY KEY,
			Name VARCHAR(50) UNIQUE NOT NULL,
			Email VARCHAR(100) UNIQUE NOT NULL,

			Password VARCHAR(100) NOT NULL,
			PasswordSalt VARCHAR(100) NOT NULL,
			PasswordHashScheme VARCHAR NOT NULL,

			Admin BOOLEAN DEFAULT FALSE,
			Phone VARCHAR DEFAULT '',
			PhoneCarrier INTEGER DEFAULT 0,

			UploadLimit_Items INTEGER DEFAULT 24000,
			ProcessingLimit_S INTEGER DEFAULT 86400,
			StorageLimit_Gb INTEGER DEFAULT 4,

			CreateTime INTEGER DEFAULT 0,
			ModifyTime INTEGER DEFAULT 0,

			UserGroup INTEGER DEFAULT 0,

			FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL
			);`)

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}
    // mysql: CREATE INDEX UserNameIndex ON User (Name);
	// postgres CREATE INDEX IF NOT EXISTS UserNameIndex ON Users (Name);
	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS UserNameIndex ON Users (Name);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}


	/**
	psql:
	CREATE TABLE IF NOT EXISTS Device
	(   Id SERIAL PRIMARY KEY,
	Name VARCHAR(100) NOT NULL,
	ApiKey VARCHAR(100) UNIQUE NOT NULL,
	Enabled BOOLEAN DEFAULT TRUE,
	Icon_PngB64 VARCHAR(512000) DEFAULT '',
	Shortname VARCHAR(100) DEFAULT '',
	Superdevice BOOL DEFAULT FALSE,
	OwnerId INTEGER,

	CanWrite BOOL DEFAULT TRUE,
	CanWriteAnywhere BOOL DEFAULT TRUE,
	UserProxy BOOL DEFAULT FALSE,

	FOREIGN KEY(OwnerId) REFERENCES Users(Id) ON DELETE CASCADE,
	UNIQUE(Name, OwnerId)
	);

	512kb icon
	**/
	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Device
		(   Id INTEGER PRIMARY KEY,
			Name STRING NOT NULL,
			ApiKey STRING UNIQUE NOT NULL,
			Enabled BOOLEAN DEFAULT TRUE,
			Icon_PngB64 STRING DEFAULT '',
			Shortname STRING DEFAULT '',
			Superdevice BOOL DEFAULT FALSE,
			OwnerId INTEGER,

			CanWrite BOOL DEFAULT TRUE,
			CanWriteAnywhere BOOL DEFAULT TRUE,
			UserProxy BOOL DEFAULT FALSE,

			FOREIGN KEY(OwnerId) REFERENCES Users(Id) ON DELETE CASCADE,
			UNIQUE(Name, OwnerId)
			);`)

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	//psql `CREATE INDEX DeviceNameIndex ON Device (Name);`
	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS DeviceNameIndex ON Device (Name);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	//psql no ine: CREATE INDEX DeviceAPIIndex ON Device (ApiKey);
	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS DeviceAPIIndex ON Device (ApiKey);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	// `CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);
	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS DeviceOwnerIndex ON Device (OwnerId);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Stream
		(   Id INTEGER PRIMARY KEY,
			Name STRING NOT NULL,
			Active BOOLEAN DEFAULT TRUE,
			Public BOOLEAN DEFAULT FALSE,
			Type VARCHAR(512) NOT NULL,
			OwnerId INTEGER,
			Ephemeral BOOL DEFAULT FALSE,
			Output BOOL DEFAULT FALSE,
			FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE,
			UNIQUE(Name, OwnerId)
			);`)

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS StreamNameIndex ON Stream (Name);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.Db.Exec(`CREATE INDEX IF NOT EXISTS StreamOwnerIndex ON Stream (OwnerId);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	return nil
}

/**
Sets up the PostgreSQL databse.
**/
func (userdb *UserDatabase) setupPostgresDatabase() error{

	log.Printf("setting up postgres db")


	_, err := userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier (
	    id integer primary key,
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);`);

	if err != nil {
		log.Printf("Error phone carrier %v", err)
		return err
	}


	_, _ = userdb.Db.Exec(`INSERT INTO PhoneCarrier VALUES (0, 'None', '')`);

    /** mysql
    CREATE TABLE IF NOT EXISTS User(   Id SERIAL PRIMARY KEY AUTO_INCREMENT, Name VARCHAR(50) UNIQUE NOT NULL, Email VARCHAR(100) UNIQUE NOT NULL, Password VARCHAR(100) NOT NULL, PasswordSalt VARCHAR(100) NOT NULL, PasswordHashScheme VARCHAR(50) NOT NULL, Admin BOOLEAN DEFAULT FALSE, Phone VARCHAR(50) DEFAULT "", PhoneCarrier INTEGER DEFAULT 0, UploadLimit_Items INTEGER DEFAULT 24000, ProcessingLimit_S INTEGER DEFAULT 86400, StorageLimit_Gb INTEGER DEFAULT 4,  FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL );
    **/
	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Users (Id SERIAL PRIMARY KEY,
			Name VARCHAR(50) UNIQUE NOT NULL,
			Email VARCHAR(100) UNIQUE NOT NULL,
			Password VARCHAR(100) NOT NULL,
			PasswordSalt VARCHAR(100) NOT NULL,
			PasswordHashScheme VARCHAR NOT NULL,
			Admin BOOLEAN DEFAULT FALSE,
			Phone VARCHAR DEFAULT '',
			PhoneCarrier INTEGER DEFAULT 0,
			UploadLimit_Items INTEGER DEFAULT 24000,
			ProcessingLimit_S INTEGER DEFAULT 86400,
			StorageLimit_Gb INTEGER DEFAULT 4,
			CreateTime INTEGER DEFAULT 0,
			ModifyTime INTEGER DEFAULT 0,
			UserGroup INTEGER DEFAULT 0,
			FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL);`)

	if err != nil {
		log.Printf("Error users %v", err)
		return err
	}


	userdb.Db.Exec(`CREATE INDEX UserNameIndex ON Users (Name);`)


	/**
	psql:


	512kb icon
	**/
	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Device
	(   Id SERIAL PRIMARY KEY,
	Name VARCHAR(100) NOT NULL,
	ApiKey VARCHAR(100) UNIQUE NOT NULL,
	Enabled BOOLEAN DEFAULT TRUE,
	Icon_PngB64 VARCHAR DEFAULT '',
	Shortname VARCHAR(100) DEFAULT '',
	Superdevice BOOL DEFAULT FALSE,
	OwnerId INTEGER,

	CanWrite BOOL DEFAULT TRUE,
	CanWriteAnywhere BOOL DEFAULT TRUE,
	UserProxy BOOL DEFAULT FALSE,

	FOREIGN KEY(OwnerId) REFERENCES Users(Id) ON DELETE CASCADE,
	UNIQUE(Name, OwnerId)
	);`)

	if err != nil {
		log.Printf("Error device %v", err)
		return err
	}


	// ignore errors b/c we don't have if not exists
	userdb.Db.Exec(`CREATE INDEX DeviceNameIndex ON Device (Name);`)
	userdb.Db.Exec(`CREATE INDEX DeviceAPIIndex ON Device (ApiKey);`)
	userdb.Db.Exec(`CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);`)

	_, err = userdb.Db.Exec(`CREATE TABLE IF NOT EXISTS Stream
	(   Id SERIAL PRIMARY KEY,
	Name VARCHAR(100) NOT NULL,
	Active BOOLEAN DEFAULT TRUE,
	Public BOOLEAN DEFAULT FALSE,
	Type VARCHAR(512) NOT NULL,
	OwnerId INTEGER,
	Ephemeral BOOL DEFAULT FALSE,
	Output BOOL DEFAULT FALSE,
	FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE,
	UNIQUE(Name, OwnerId)
	);`)

	if err != nil {
		log.Printf("Error stream %v", err)
		return err
	}


	userdb.Db.Exec(`CREATE INDEX StreamNameIndex ON Stream (Name);`)
	userdb.Db.Exec(`CREATE INDEX StreamOwnerIndex ON Stream (OwnerId);`)


	// setup our statements to work with postgres
	CREATE_PHONE_CARRIER_STMT = `INSERT INTO PhoneCarrier (Name, EmailDomain) VALUES ($1,$2);`
	SELECT_PHONE_CARRIER_BY_ID_STMT = "SELECT * FROM PhoneCarrier WHERE Id = $1 LIMIT 1"
	UPDATE_PHONE_CARRIER_STMT = `UPDATE PhoneCarrier SET Name=$1, EmailDomain=$2 WHERE Id = $3;`
	DELETE_PHONE_CARRIER_BY_ID_STMT = `DELETE FROM PhoneCarrier WHERE Id = $1;`
	CREATE_DEVICE_STMT = `INSERT INTO Device
	    (	Name,
	        ApiKey,
	        Icon_PngB64,
	        OwnerId)
	        VALUES ($1,$2,$3,$4)`

	SELECT_DEVICE_BY_USER_ID_STMT = "SELECT * FROM Device WHERE OwnerId = $1"
	SELECT_DEVICE_BY_ID_STMT = "SELECT * FROM Device WHERE Id = $1 LIMIT 1"
	SELECT_DEVICE_BY_API_KEY_STMT = "SELECT * FROM Device WHERE ApiKey = $1 LIMIT 1"
	UPDATE_DEVICE_STMT = `UPDATE Device SET
	    Name = $1, ApiKey = $2, Enabled = $3,
	    Icon_PngB64 = $4, Shortname = $5, Superdevice = $6,
	    OwnerId = $7, CanWrite = $8, CanWriteAnywhere = $9, UserProxy = $10 WHERE Id = $11;`
	DELETE_DEVICE_BY_ID_STMT = `DELETE FROM Device WHERE Id = $1;`
	CREATE_STREAM_STMT = `INSERT INTO Stream
	    (	Name,
	        Type,
	        OwnerId) VALUES ($1,$2,$3);`
	SELECT_STREAM_BY_ID_STMT = "SELECT * FROM Stream WHERE Id = $1 LIMIT 1"
	SELECT_STREAM_BY_DEVICE_STMT = "SELECT * FROM Stream WHERE OwnerId = $1"
	UPDATE_STREAM_STMT = `UPDATE Stream SET
	    Name = $1,
	    Active = $2,
	    Public = $3,
	    Type = $4,
	    OwnerId = $5,
	    Ephemeral = $6,
	    Output = $7
	    WHERE Id = $8;`
	DELETE_STREAM_BY_ID_STMT = `DELETE FROM Stream WHERE Id = $1;`
	CREATE_USER_STMT = `INSERT INTO Users (
	    Name,
	    Email,
	    Password,
	    PasswordSalt,
	    PasswordHashScheme,
	    CreateTime) VALUES ($1,$2,$3,$4,$5,$6);`
	SELECT_USER_BY_EMAIL_STMT = "SELECT * FROM Users WHERE Email = $1 LIMIT 1"
	SELECT_USER_BY_NAME_STMT = "SELECT * FROM Users WHERE Name = $1 LIMIT 1"
	SELECT_USER_BY_ID_STMT = "SELECT * FROM Users WHERE Id = $1 LIMIT 1"
	SELECT_ALL_USERS_STMT = "SELECT * FROM Users"
	SELECT_OWNER_OF_STREAM_BY_ID_STMT = `SELECT u.*
	                              FROM Users u, Stream s, Device d
	                              WHERE s.Id = $1
	                                AND d.Id = s.OwnerId
	                                AND u.Id = d.OwnerId
	                              LIMIT 1;`
	UPDATE_USER_STMT = `UPDATE Users SET
	                Name=$1, Email=$2, Password=$3, PasswordSalt=$4, PasswordHashScheme=$5,
	                Admin=$6, Phone=$7, PhoneCarrier=$8, UploadLimit_Items=$9,
	                ProcessingLimit_S=$10, StorageLimit_Gb=$11, CreateTime = $12, ModifyTime = $13,
	                UserGroup = $14 WHERE Id = $15;`
	DELETE_USER_BY_ID_STMT = `DELETE FROM Users WHERE Id = $1;`
	READ_DEVICE_BY_USER_AND_NAME = "SELECT * FROM Device WHERE OwnerId = $1 AND Name = $2 LIMIT 1"
	READ_STREAM_BY_DEVICE_AND_NAME = "SELECT * FROM Stream WHERE OwnerId = $1 AND Name = $2 LIMIT 1"

	return nil
}
