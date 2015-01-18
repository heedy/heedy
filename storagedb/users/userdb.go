package main

/**

stream = {
name: string
id: int
active:
description: string
public: bool
device: <deviceid>
schema: string
defaults: string
}
**/

//package datastore


import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"github.com/nu7hatch/gouuid"
//	"crypto/sha512"
	"errors"
//	"encoding/hex"
)


var DB_DRIVER string
var db *sql.DB // the database



type User struct {
	user_id int
	user_name string
	user_email string
	user_pass string
	user_pass_salt string
	user_pass_hash_scheme string
	user_admin bool
	user_phone string
	user_phone_carrier string // phone carrier string
	user_upload_limit int // upload limit in items/day
	user_process_limit int // processing limit in seconds/day
	user_storage_limit int // storage limit in GB
}

type PhoneCarrier struct {
	carrier_id int
	carrier_name string
	carrier_email_suffix string
}


type Device struct {
	device_id int
	device_name string
	device_api_key string
	device_enabled bool
	device_icon *[]byte // a png image
	device_shortname string
	device_superdevice bool
	device_owner int // a user
}

type Stream struct {
	stream_id int
	stream_name string
	stream_active bool
	stream_public bool
	stream_schema string
	stream_defaults string
	stream_owner int
}


func InitConnection() (err error) {
	//sql.Register(DB_DRIVER, &sqlite3.SQLiteDriver{})


	return nil
}


/**
Sets up the SQLITE databse.
**/
func setupDatabase() {

	_, err := db.Exec("PRAGMA foreign_keys = ON;")

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier
		(	carrier_id INTEGER PRIMARY KEY,
			carrier_name STRING UNIQUE NOT NULL,
			carrier_email_suffix STRING UNIQUE NOT NULL)`);

	if err != nil {
		log.Fatal(err)
	}

	_, _ = db.Exec(`INSERT INTO PhoneCarrier VALUES (0, "None", "")`);


	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS User
		(   user_id INTEGER PRIMARY KEY,
			user_name STRING UNIQUE NOT NULL,
			user_email STRING UNIQUE NOT NULL,
			user_pass STRING NOT NULL,
			user_pass_salt STRING NOT NULL,
			user_pass_hash_scheme STRING NOT NULL,
			user_admin BOOLEAN DEFAULT FALSE,
			user_phone STRING DEFAULT "",
			user_phone_carrier INTEGER DEFAULT 0,
			user_upload_limit INTEGER DEFAULT 24000,
			user_process_limit INTEGER DEFAULT 86400,
			user_storage_limit INTEGER DEFAULT 4,

			FOREIGN KEY(user_phone_carrier) REFERENCES PhoneCarrier(carrier_id) ON DELETE SET NULL
			);`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS UserNameIndex ON User (user_name);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Device
		(   device_id INTEGER PRIMARY KEY,
			device_name STRING NOT NULL,
			device_api_key STRING NOT NULL,
			device_enabled BOOLEAN DEFAULT TRUE,
			device_icon BLOB,
			device_shortname STRING DEFAULT "",
			device_superdevice BOOL DEFAULT FALSE,
			device_owner INTEGER,
			FOREIGN KEY(device_owner) REFERENCES User(user_id) ON DELETE CASCADE
			);`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS DeviceNameIndex ON Device (device_name);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS DeviceAPIIndex ON Device (device_api_key);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS DeviceOwnerIndex ON Device (device_owner);`)
	if err != nil {
		log.Fatal(err)
	}


	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Stream
		(   stream_id INTEGER PRIMARY KEY,
			stream_name STRING NOT NULL,
			stream_active BOOLEAN DEFAULT TRUE,
			stream_public BOOLEAN DEFAULT FALSE,
			stream_schema STRING NOT NULL,
			stream_defaults STRING NOT NULL,
			stream_owner INTEGER,
			FOREIGN KEY(stream_owner) REFERENCES Device(device_id) ON DELETE CASCADE
			);`)


	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS StreamNameIndex ON Stream (stream_name);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS StreamOwnerIndex ON Stream (stream_owner);`)
	if err != nil {
		log.Fatal(err)
	}


}




func CreateUser(user_name, user_email, user_pass string) (err error) {

	// Ensure we don't have someone with the same email or name
	usr, err := ReadUserByEmail(user_email)
	if(err == nil){
		return errors.New("A user already exists with this email")
	}

	usr, err = ReadUserByName(user_name)
	if(usr == nil){
		return errors.New("A user already exists with this name")
	}

	// Now do the insert

	user_pass_salt, _ := uuid.NewV4()
	user_hash_scheme := "SHA512"
	saltedpass := user_pass + user_pass_salt.String()

	// TODO decide and do encryption here // []byte(sha512.Sum512([]byte(saltedpass)))?
	dbpass := saltedpass

	_, err = db.Exec(`INSERT INTO User (
		user_name,
		user_email,
		user_pass,
		user_pass_salt,
		user_pass_hash_scheme) VALUES (?,?,?,?,?);`,
		user_name,
		user_email,
		dbpass,
		user_pass_salt.String(),
		user_hash_scheme)

	return err
}

func constructUserFromRow(rows *sql.Rows) (*User, error){
	u := new(User)

	for rows.Next() {
		err := rows.Scan(
					&u.user_id,
					&u.user_name,
					&u.user_email,
					&u.user_pass,
					&u.user_pass_salt,
					&u.user_pass_hash_scheme,
					&u.user_admin,
					&u.user_phone,
					&u.user_phone_carrier,
					&u.user_upload_limit,
					&u.user_process_limit,
					&u.user_storage_limit)
		if err != nil {
			return u, err
		}

		return u, nil
	}

	return u, errors.New("No user supplied")
}

func ReadUserByEmail(user_email string) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_email = ?", user_email)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructUserFromRow(rows)
}


func ReadUserByName(user_name string) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_name = ?", user_name)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructUserFromRow(rows)
}

func ReadUserById(user_id int) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_id = ?", user_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructUserFromRow(rows)
}

func UpdateUser(user User) (error) {
	_, err := db.Exec(`UPDATE User SET
		user_name=?, user_email=?, user_pass=?, user_pass_salt=?, user_pass_hash_scheme=?,
			user_admin=?, user_phone=?, user_phone_carrier=?, user_upload_limit=?,
			user_process_limit=?, user_storage_limit=? WHERE user_id = ?;`,
			user.user_name,
			user.user_email,
			user.user_pass,
			user.user_pass_salt,
			user.user_pass_hash_scheme,
			user.user_admin,
			user.user_phone,
			user.user_phone_carrier,
			user.user_upload_limit,
			user.user_process_limit,
			user.user_storage_limit,
			user.user_id);
	return err
}

func DeleteUser(user User) (error) {
	_, err := db.Exec(`DELETE FROM User WHERE user_id = ?;`, user.user_id );
	return err
}





/**

type PhoneCarrier struct {
	carrier_id int
	carrier_name string
	carrier_email_suffix string
}


type Device struct {
	device_id int
	device_name string
	device_api_key string
	device_enabled bool
	device_icon string // a base64 png image
	device_shortname string
	device_superdevice bool
	device_owner int // a user
}

type Stream struct {
	stream_id int
	stream_name string
	stream_active bool
	stream_public bool
	stream_schema string
	stream_defaults string
	stream_owner int
}


**/



func main() {
	log.Print("testing startup")
	var err error
	db, err = sql.Open("sqlite3", "users.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Print("closing down")
	}

	setupDatabase()

	err = CreateUser("Joseph", "joseph@josephlewis.net", "password")
	if(err != nil){
		log.Print("Cannot create user " + err.Error())
	}

	u, err := ReadUserByName("Joseph")
	if(err != nil){
		log.Print("Cannot read user"  + err.Error())
	}


	log.Print(u.user_email)

	log.Print("closing down")

}
