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
)

type DRIVERSTR string;

const(
	SQLITE3 DRIVERSTR = "sqlite3"
	POSTGRES DRIVERSTR = "postgres"
)

// The main UserDatabase type
type UserDatabase struct {
	driverstr DRIVERSTR
	filepath string
	port int
	db *sql.DB
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


func (n *UserDatabase) InitUserDatabase(ds DRIVERSTR, cxn string) error {
	n.driverstr = ds
	n.filepath = cxn

	var err error
	n.db, err = sql.Open(string(n.driverstr), n.filepath)
	if err != nil {
		return err
	}

	err = n.db.Ping()
	if err != nil {
		return err
	}

	switch ds {
		case SQLITE3:
			if err := n.setupSqliteDatabase(); err != nil {
				return err
			}
		case POSTGRES:
			if err := n.setupPostgresDatabase(); err != nil {
				return err
			}
		default:
			return errors.New("Illegal driver string")
	}

	return nil
}

func (userdb *UserDatabase) UnderlyingDb() *sql.DB {
	return userdb.db
}




/**
Sets up the SQLITE databse.
**/
func (userdb *UserDatabase) setupSqliteDatabase() error{

	log.Printf("setting up squlite db")

	_, err := userdb.db.Exec("PRAGMA foreign_keys = ON;")

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier (
	    id integer primary key,
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);`);

	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, _ = userdb.db.Exec(`INSERT INTO PhoneCarrier VALUES (0, 'None', '')`);

    /** mysql
    CREATE TABLE IF NOT EXISTS User(   Id SERIAL PRIMARY KEY AUTO_INCREMENT, Name VARCHAR(50) UNIQUE NOT NULL, Email VARCHAR(100) UNIQUE NOT NULL, Password VARCHAR(100) NOT NULL, PasswordSalt VARCHAR(100) NOT NULL, PasswordHashScheme VARCHAR(50) NOT NULL, Admin BOOLEAN DEFAULT FALSE, Phone VARCHAR(50) DEFAULT "", PhoneCarrier INTEGER DEFAULT 0, UploadLimit_Items INTEGER DEFAULT 24000, ProcessingLimit_S INTEGER DEFAULT 86400, StorageLimit_Gb INTEGER DEFAULT 4,  FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL );
    **/
	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Users (Id INTEGER PRIMARY KEY,
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
	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS UserNameIndex ON Users (Name);`)
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
	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Device
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
	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceNameIndex ON Device (Name);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	//psql no ine: CREATE INDEX DeviceAPIIndex ON Device (ApiKey);
	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceAPIIndex ON Device (ApiKey);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	// `CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);
	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceOwnerIndex ON Device (OwnerId);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Stream
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

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS StreamNameIndex ON Stream (Name);`)
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS StreamOwnerIndex ON Stream (OwnerId);`)
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


	_, err := userdb.db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier (
	    id integer primary key,
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);`);

	if err != nil {
		log.Printf("Error phone carrier %v", err)
		return err
	}


	_, _ = userdb.db.Exec(`INSERT INTO PhoneCarrier VALUES (0, 'None', '')`);

    /** mysql
    CREATE TABLE IF NOT EXISTS User(   Id SERIAL PRIMARY KEY AUTO_INCREMENT, Name VARCHAR(50) UNIQUE NOT NULL, Email VARCHAR(100) UNIQUE NOT NULL, Password VARCHAR(100) NOT NULL, PasswordSalt VARCHAR(100) NOT NULL, PasswordHashScheme VARCHAR(50) NOT NULL, Admin BOOLEAN DEFAULT FALSE, Phone VARCHAR(50) DEFAULT "", PhoneCarrier INTEGER DEFAULT 0, UploadLimit_Items INTEGER DEFAULT 24000, ProcessingLimit_S INTEGER DEFAULT 86400, StorageLimit_Gb INTEGER DEFAULT 4,  FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL );
    **/
	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Users (Id SERIAL PRIMARY KEY,
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


	userdb.db.Exec(`CREATE INDEX UserNameIndex ON Users (Name);`)


	/**
	psql:


	512kb icon
	**/
	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Device
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
	);`)

	if err != nil {
		log.Printf("Error device %v", err)
		return err
	}


	// ignore errors b/c we don't have if not exists
	userdb.db.Exec(`CREATE INDEX DeviceNameIndex ON Device (Name);`)
	userdb.db.Exec(`CREATE INDEX DeviceAPIIndex ON Device (ApiKey);`)
	userdb.db.Exec(`CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);`)

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Stream
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


	userdb.db.Exec(`CREATE INDEX StreamNameIndex ON Stream (Name);`)
	userdb.db.Exec(`CREATE INDEX StreamOwnerIndex ON Stream (OwnerId);`)

	return nil
}
