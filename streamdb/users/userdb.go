package main

//package datastore

// TODO This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"github.com/nu7hatch/gouuid"
//	"crypto/sha512"
	"errors"
//	"encoding/hex"
)


const(
	// A black and qhite question mark
	DEFAULT_ICON =`iVBORw0KGgoAAAANSUhEUgAAAEAAAABAAQMAAACQp+OdAAAABlBMVEUAA
	AAAAAClZ7nPAAAAAXRSTlMAQObYZgAAAAFiS0dEAIgFHUgAAAAJcEhZcwAACxMAAAsTAQCanBgAAACVS
	URBVCjPjdGxDcQgDAVQRxSUHoFRMtoxWkZhBEoKK/+IsaNc0ElQxE8K3xhBtLa4Gj4YNQBFEYHxjwFRJ
	OBU7AAsZOgVWSEJR68bajSUoOjfoK07NkP+h/jAiI8g2WgGdqRx+jVa/r0P2cx9EPE2zduUVxv2NHs6n
	Q6Z0BZQaX3F4/0od3xvE2TCtOeOs12UQl6c5Quj42jQ5zt8GQAAAABJRU5ErkJggg==`
)

var (
 	DB_DRIVER string
	db *sql.DB // the database
)



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
	carrier_email_domain string
}


type Device struct {
	device_id int
	device_name string
	device_api_key string
	device_enabled bool
	device_icon string // a png image in base64
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
			carrier_email_domain STRING UNIQUE NOT NULL)`);

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
			device_icon STRING DEFAULT "",
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
	if(usr != nil){
		return errors.New("A user already exists with this email")
	}

	usr, err = ReadUserByName(user_name)
	if(usr != nil){
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
		return u, err
	}

	return nil, errors.New("No user supplied")
}

func ReadUserByEmail(user_email string) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_email = ? LIMIT 1", user_email)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructUserFromRow(rows)
}


func ReadUserByName(user_name string) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_name = ? LIMIT 1", user_name)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructUserFromRow(rows)
}

func ReadUserById(user_id int) (*User, error){
	rows, err := db.Query("SELECT * FROM User WHERE user_id = ? LIMIT 1", user_id)

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


func CreatePhoneCarrier(carrier_name, carrier_email_domain string) (error) {

	_, err := db.Exec(`INSERT INTO PhoneCarrier (
		carrier_name,
		carrier_email_domain) VALUES (?,?);`,
		carrier_name,
		carrier_email_domain)

	return err
}

func constructPhoneCarrierFromRow(rows *sql.Rows) (*PhoneCarrier, error){
	u := new(PhoneCarrier)

	for rows.Next() {
		err := rows.Scan(
			&u.carrier_id,
			&u.carrier_name,
			&u.carrier_email_domain)

		if err != nil {
			return u, err
		}

		return u, nil
	}

	return u, errors.New("No carrier supplied")
}

func ReadPhoneCarrierById(carrier_id int) (*PhoneCarrier, error) {
	rows, err := db.Query("SELECT * FROM PhoneCarrier WHERE carrier_id = ? LIMIT 1", carrier_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructPhoneCarrierFromRow(rows)
}

func UpdatePhoneCarrier(carrier *PhoneCarrier) (error) {
	_, err := db.Exec(`UPDATE PhoneCarrier SET
		carrier_name=?, carrier_email_domain=? WHERE carrier_id = ?;`,
		carrier.carrier_name,
		carrier.carrier_email_domain,
		carrier.carrier_id);
	return err
}


func DeletePhoneCarrier(carrier *PhoneCarrier) (error) {
	_, err := db.Exec(`DELETE FROM PhoneCarrier WHERE carrier_id = ?;`, carrier.carrier_id );
	return err
}


func CreateDevice(device_name string, device_owner *User) (error) {
	device_api_key, _ := uuid.NewV4()

	_, err := db.Exec(`INSERT INTO Device
		(	device_name,
			device_api_key,
			device_icon,
			device_owner)
		VALUES (?,?,?,?,?)`,
		device_name, device_api_key.String(), DEFAULT_ICON, device_owner.user_id)
	return err
}

func constructDeviceFromRow(rows *sql.Rows) (*Device, error) {
	u := new(Device)

	for rows.Next() {
		err := rows.Scan(
			&u.device_id,
			&u.device_name,
			&u.device_api_key,
			&u.device_enabled,
			&u.device_icon,
			&u.device_shortname,
			&u.device_superdevice,
			&u.device_owner)

			return u, err
	}

	return u, errors.New("No carrier supplied")
}

func ReadDeviceById(device_id int) (*Device, error) {
	rows, err := db.Query("SELECT * FROM Device WHERE device_id = ? LIMIT 1", device_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return constructDeviceFromRow(rows)

}

func UpdateDevice(device *Device) (error) {
	_, err := db.Exec(`UPDATE Device SET
			device_name = ?, device_api_key = ?, device_enabled = ?,
			device_icon = ?, device_shortname = ?, device_superdevice = ?,
			device_owner = ? WHERE device_id = ?;`,
			device.device_name,
			device.device_api_key,
			device.device_enabled,
			device.device_icon,
			device.device_shortname,
			device.device_superdevice,
			device.device_owner,
			device.device_id)

	return err
}


func DeleteDevice(device *Device) (error) {
	_, err := db.Exec(`DELETE FROM Device WHERE device_id = ?;`, device.device_id );
	return err
}




func CreateStream(stream_name, stream_schema, stream_defaults string, owner *Device) (error) {
	_, err := db.Exec(`INSERT INTO Stream
		(	stream_name,
			stream_schema,
			stream_defaults,
			stream_owner) VALUES (?,?,?,?);`,
			stream_name, stream_schema, stream_defaults, owner.device_id)
	return err
}

func constructDevicesFromRows(rows *sql.Rows) ([]*Device, error) {
	out := []*Device{}

	for rows.Next() {
		u := new(Device)
		err := rows.Scan(
			&u.device_id,
			&u.device_name,
			&u.device_api_key,
			&u.device_enabled,
			&u.device_icon,
			&u.device_shortname,
			&u.device_superdevice,
			&u.device_owner)

		out = append(out, u)

		if(err != nil) {
			return out, err
		}
	}

	return out, nil
}

func ReadStreamById(id int) (*Device, error) {
	rows, err := db.Query("SELECT * FROM Stream WHERE stream_id = ? LIMIT 1", id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	devices, err := constructDevicesFromRows(rows)

	if(len(devices) != 1) {
		return nil, errors.New("Wrong number of streams returned")
	}

	return devices[0], nil
}

func UpdateStream(stream *Stream) (error) {
	_, err := db.Exec(`UPDATE Stream SET
		stream_name = ?,
		stream_active = ?,
		stream_public = ?,
		stream_schema = ?,
		stream_defaults = ?,
		stream_owner = ? WHERE stream_id = ?`,
		stream.stream_name,
		stream.stream_active,
		stream.stream_public,
		stream.stream_schema,
		stream.stream_defaults,
		stream.stream_owner,
		stream.stream_id)

	return err
}

func DeleteStream(stream *Stream) (error) {
	_, err := db.Exec(`DELETE FROM Stream WHERE stream_id = ?;`, stream.stream_id );
	return err
}

/**
func init() {
	if user == "" {
		log.Fatal("$USER not set")
	}
	if home == "" {
		home = "/home/" + user
	}
	if gopath == "" {
		gopath = home + "/go"
	}
	// gopath may be overridden by --gopath flag on command line.
	flag.StringVar(&gopath, "gopath", gopath, "override default GOPATH")
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
		log.Print("Cannot read user "  + err.Error())
	}

	if(u != nil) {
		log.Print(u.user_email)
	}
	log.Print("closing down")

}
