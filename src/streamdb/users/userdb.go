// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"errors"

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

// The main UserDatabase type
type UserDatabase struct {
	driverstr string
	filepath string
	port int
	db *sql.DB
}

// Opens the database for operations
func (userdb *UserDatabase) open() (error){
	var err error
	userdb.db, err = sql.Open(userdb.driverstr, userdb.filepath)
	if err != nil {
		return err
	}

	err = userdb.db.Ping()
	if err != nil {
		return err
	}

	return userdb.setupDatabase()
}

func NewSqliteUserDatabase(path string) (*UserDatabase, error) {
	n := new(UserDatabase)

	n.driverstr = "sqlite3"
	n.filepath = path

	if err := n.open(); err != nil {
		return nil, err
	}

	if err := n.setupDatabase(); err != nil {
		return nil, err
	}

	return n, nil
}

func NewMariaUserDatabase(url, dbname string, port int) (*UserDatabase, error) {

	return nil, nil
}




/**
Sets up the SQLITE databse.
**/
func (userdb *UserDatabase) setupDatabase() error{

	_, err := userdb.db.Exec("PRAGMA foreign_keys = ON;")

	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier (
	    id integer primary key,
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);`);

	if err != nil {
		return err
	}

	_, _ = userdb.db.Exec(`INSERT INTO PhoneCarrier VALUES (0, "None", "")`);

    /** mysql
    CREATE TABLE IF NOT EXISTS User(   Id INTEGER PRIMARY KEY AUTO_INCREMENT, Name VARCHAR(50) UNIQUE NOT NULL, Email VARCHAR(100) UNIQUE NOT NULL, Password VARCHAR(100) NOT NULL, PasswordSalt VARCHAR(100) NOT NULL, PasswordHashScheme VARCHAR(50) NOT NULL, Admin BOOLEAN DEFAULT FALSE, Phone VARCHAR(50) DEFAULT "", PhoneCarrier INTEGER DEFAULT 0, UploadLimit_Items INTEGER DEFAULT 24000, ProcessingLimit_S INTEGER DEFAULT 86400, StorageLimit_Gb INTEGER DEFAULT 4,  FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL );
    **/
	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS User
		(   Id INTEGER PRIMARY KEY,
			Name VARCHAR(50) UNIQUE NOT NULL,
			Email VARCHAR(100) UNIQUE NOT NULL,

			Password VARCHAR(100) NOT NULL,
			PasswordSalt VARCHAR(100) NOT NULL,
			PasswordHashScheme VARCHAR NOT NULL,

			Admin BOOLEAN DEFAULT FALSE,
			Phone VARCHAR DEFAULT "",
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
		return err
	}
    // mysql: CREATE INDEX UserNameIndex ON User (Name);
	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS UserNameIndex ON User (Name);`)
	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Device
		(   Id INTEGER PRIMARY KEY,
			Name STRING NOT NULL,
			ApiKey STRING UNIQUE NOT NULL,
			Enabled BOOLEAN DEFAULT TRUE,
			Icon_PngB64 STRING DEFAULT "",
			Shortname STRING DEFAULT "",
			Superdevice BOOL DEFAULT FALSE,
			OwnerId INTEGER,

			CanWrite BOOL DEFAULT TRUE,
			CanWriteAnywhere BOOL DEFAULT TRUE,
			UserProxy BOOL DEFAULT FALSE,

			FOREIGN KEY(OwnerId) REFERENCES User(Id) ON DELETE CASCADE,
			UNIQUE(Name, OwnerId)
			);`)

	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceNameIndex ON Device (Name);`)
	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceAPIIndex ON Device (ApiKey);`)
	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS DeviceOwnerIndex ON Device (OwnerId);`)
	if err != nil {
		return err
	}


	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS Stream
		(   Id INTEGER PRIMARY KEY,
			Name STRING NOT NULL,
			Active BOOLEAN DEFAULT TRUE,
			Public BOOLEAN DEFAULT FALSE,
			Type STRING NOT NULL,
			OwnerId INTEGER,
			Ephemeral BOOL DEFAULT FALSE,
			Output BOOL DEFAULT FALSE,
			FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE,
			UNIQUE(Name, OwnerId)
			);`)

	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS StreamNameIndex ON Stream (Name);`)
	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE INDEX IF NOT EXISTS StreamOwnerIndex ON Stream (OwnerId);`)
	if err != nil {
		return err
	}

	return nil
}
