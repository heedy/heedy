// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
	"errors"
	"crypto/sha512"
	"encoding/hex"
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

// The main UserDatabase type
type UserDatabase struct {
	driverstr string
	filepath string
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




/**
Sets up the SQLITE databse.
**/
func (userdb *UserDatabase) setupDatabase() error{

	_, err := userdb.db.Exec("PRAGMA foreign_keys = ON;")

	if err != nil {
		return err
	}

	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS PhoneCarrier
		(	Id INTEGER PRIMARY KEY,
			Name STRING UNIQUE NOT NULL,
			EmailDomain STRING UNIQUE NOT NULL)`);

	if err != nil {
		return err
	}

	_, _ = userdb.db.Exec(`INSERT INTO PhoneCarrier VALUES (0, "None", "")`);


	_, err = userdb.db.Exec(`CREATE TABLE IF NOT EXISTS User
		(   Id INTEGER PRIMARY KEY,
			Name STRING UNIQUE NOT NULL,
			Email STRING UNIQUE NOT NULL,
			Password STRING NOT NULL,
			PasswordSalt STRING NOT NULL,
			PasswordHashScheme STRING NOT NULL,
			Admin BOOLEAN DEFAULT FALSE,
			Phone STRING DEFAULT "",
			PhoneCarrier INTEGER DEFAULT 0,
			UploadLimit_Items INTEGER DEFAULT 24000,
			ProcessingLimit_S INTEGER DEFAULT 86400,
			StorageLimit_Gb INTEGER DEFAULT 4,

			FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL
			);`)

	if err != nil {
		return err
	}

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
			Schema_Json STRING NOT NULL,
			Defaults_Json STRING NOT NULL,
			OwnerId INTEGER,
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

// calcHash calculates the user hash for the given password, salt and hashing
// scheme
func calcHash(password, salt, scheme string) (string) {
	switch scheme {
		// We switch over hashes here so if we need to upgrade in the future,
		// it is easy.
		case "SHA512":
			saltedpass := password + salt

			hasher := sha512.New()
			hasher.Write([]byte(saltedpass))
			return hex.EncodeToString(hasher.Sum(nil))
		default:
			return calcHash(password, salt, "SHA512")
	}
}

// ValidateUser checks to see if a user going by the username or email
// matches the given password, returns true if it does false if it does not
func (userdb *UserDatabase) ValidateUser(UsernameOrEmail, Password string) (bool, *User) {
	var usr *User
	var err error

	usr, err = userdb.ReadUserByName(UsernameOrEmail)
	if err != nil {
		log.Print(err)
	}
	if usr != nil {
		goto gotuser
	}

	usr, err = userdb.ReadUserByEmail(UsernameOrEmail)
	if err != nil {
		log.Print(err)
	}
	if usr != nil {
		goto gotuser
	}

gotuser:
	log.Printf("User: %v", usr)
	if usr != nil && calcHash(Password, usr.PasswordSalt, usr.PasswordHashScheme) == usr.Password {
		return true, usr
	} else {
		return false, nil
	}
}


// CreateUser creates a user given the user's credentials.
// If a user already exists with the given credentials, an error is thrown.
func (userdb *UserDatabase) CreateUser(Name, Email, Password string) (id int64, err error) {

	// Ensure we don't have someone with the same email or name
	usr, err := userdb.ReadUserByEmail(Email)
	if(usr != nil){
		return -1, ERR_EMAIL_EXISTS
	}

	usr, err = userdb.ReadUserByName(Name)
	if(usr != nil){
		return -1, ERR_USERNAME_EXISTS
	}

	PasswordSalt, _ := uuid.NewV4()
	dbpass := calcHash(Password, PasswordSalt.String(), DEFAULT_PASSWORD_HASH)


	// Note that golang uses utf8 strings converted to bytes first, so the hashes
	// may not match up with hash generators found online!
	//log.Print("passwordtest ", saltedpass, []byte(saltedpass), dbpass)

	res, err := userdb.db.Exec(`INSERT INTO User (
		Name,
		Email,
		Password,
		PasswordSalt,
		PasswordHashScheme) VALUES (?,?,?,?,?);`,
		Name,
		Email,
		dbpass,
		PasswordSalt.String(),
		DEFAULT_PASSWORD_HASH)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

// constructUserFromRow converts a sql.Rows object to a single user
func constructUserFromRow(rows *sql.Rows, err error) (*User, error){
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users, err := constructUsersFromRows(rows)

	if err == nil && len(users) > 0 {
		return users[0], err
	}

	return nil, errors.New("No user supplied")
}

func constructUsersFromRows(rows *sql.Rows) ([]*User, error){
	out := []*User{}

	if rows == nil {
		return out, ERR_INVALID_PTR
	}


	for rows.Next() {
		u :=  new(User)
		err := rows.Scan(&u.Id,
			&u.Name,
			&u.Email,
			&u.Password,
			&u.PasswordSalt,
			&u.PasswordHashScheme,
			&u.Admin,
			&u.Phone,
			&u.PhoneCarrier,
			&u.UploadLimit_Items,
			&u.ProcessingLimit_S,
			&u.StorageLimit_Gb)

		if err != nil {
			return out, err
		}

		out = append(out, u)
	}

	return out, nil
}

// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadUserByEmail(Email string) (*User, error){
	rows, err := userdb.db.Query("SELECT * FROM User WHERE Email = ? LIMIT 1", Email)
	return constructUserFromRow(rows, err)
}

// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *UserDatabase) ReadUserByName(Name string) (*User, error){
	rows, err := userdb.db.Query("SELECT * FROM User WHERE Name = ? LIMIT 1", Name)
	return constructUserFromRow(rows, err)
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *UserDatabase) ReadUserById(Id int64) (*User, error){
	rows, err := userdb.db.Query("SELECT * FROM User WHERE Id = ? LIMIT 1", Id)
	return constructUserFromRow(rows, err)
}

func (userdb *UserDatabase) ReadAllUsers() ([]*User, error) {
	rows, err := userdb.db.Query("SELECT * FROM User")

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return constructUsersFromRows(rows)
}


// UpdateUser updates the user with the given id in the database using the
// information provided in the user struct.
func (userdb *UserDatabase) UpdateUser(user *User) (error) {

	if user == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.db.Exec(`UPDATE User SET
		Name=?, Email=?, Password=?, PasswordSalt=?, PasswordHashScheme=?,
			Admin=?, Phone=?, PhoneCarrier=?, UploadLimit_Items=?,
			ProcessingLimit_S=?, StorageLimit_Gb=? WHERE Id = ?;`,
			user.Name,
			user.Email,
			user.Password,
			user.PasswordSalt,
			user.PasswordHashScheme,
			user.Admin,
			user.Phone,
			user.PhoneCarrier,
			user.UploadLimit_Items,
			user.ProcessingLimit_S,
			user.StorageLimit_Gb,
			user.Id);
	return err
}

// DeleteUser removes a user from the database
func (userdb *UserDatabase) DeleteUser(id int64) (error) {
	_, err := userdb.db.Exec(`DELETE FROM User WHERE Id = ?;`, id );
	return err
}

// CreatePhoneCarrier creates a phone carrier from the carrier name and
// the SMS email domain they provide, for Example "Tmobile US", "tmomail.net"
func (userdb *UserDatabase) CreatePhoneCarrier(Name, EmailDomain string) (int64, error) {

	res, err := userdb.db.Exec(`INSERT INTO PhoneCarrier (Name, EmailDomain)
						 VALUES (?,?);`,
		Name,
		EmailDomain)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}


// constructPhoneCarrierFromRow creates a single PhoneCarrier instance from
// the given rows.
func constructPhoneCarrierFromRow(rows *sql.Rows, err error) (*PhoneCarrier, error) {
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result, err := constructPhoneCarriersFromRows(rows)

	if err == nil && len(result) > 0 {
		return result[0], err
	}

	return nil, errors.New("No carrier supplied")
}

// constructPhoneCarriersFromRows constructs a series of phone carriers
func constructPhoneCarriersFromRows(rows *sql.Rows) ([]*PhoneCarrier, error) {
	out := []*PhoneCarrier{}

	if rows == nil {
		return out, ERR_INVALID_PTR
	}


	for rows.Next() {
		u := new(PhoneCarrier)
		err := rows.Scan(&u.Id, &u.Name, &u.EmailDomain)

		if err != nil {
			return out, err
		}

		out = append(out, u)
	}

	return out, nil
}

// ReadPhoneCarrierById selects a phone carrier from the database given its ID
func (userdb *UserDatabase) ReadPhoneCarrierById(Id int64) (*PhoneCarrier, error) {
	rows, err := userdb.db.Query("SELECT * FROM PhoneCarrier WHERE Id = ? LIMIT 1", Id)
	return constructPhoneCarrierFromRow(rows, err)
}

func (userdb *UserDatabase) ReadAllPhoneCarriers() ([]*PhoneCarrier, error) {
	rows, err := userdb.db.Query("SELECT * FROM PhoneCarrier")

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return constructPhoneCarriersFromRows(rows)
}

// UpdatePhoneCarrier updates the database's phone carrier data with that of the
// struct provided.
func (userdb *UserDatabase) UpdatePhoneCarrier(carrier *PhoneCarrier) (error) {
	if carrier == nil {
		return ERR_INVALID_PTR
	}


	_, err := userdb.db.Exec(`UPDATE PhoneCarrier SET
		Name=?, EmailDomain=? WHERE Id = ?;`,
		carrier.Name,
		carrier.EmailDomain,
		carrier.Id);
	return err
}

// DeletePhoneCarrier removes a phone carrier from the database, this will set
// all users carrier with this phone carrier as a foreign key to NULL
func (userdb *UserDatabase) DeletePhoneCarrier(carrierId int64) (error) {
	_, err := userdb.db.Exec(`DELETE FROM PhoneCarrier WHERE Id = ?;`, carrierId );
	return err
}

// CreateDevice adds a device to the system given its owner and name.
// returns the last inserted id
func (userdb *UserDatabase) CreateDevice(Name string, OwnerId *User) (int64, error) {
	// guards
	if OwnerId == nil {
		return -1, ERR_INVALID_PTR
	}

	ApiKey, _ := uuid.NewV4()

	res, err := userdb.db.Exec(`INSERT INTO Device
		(	Name,
			ApiKey,
			Icon_PngB64,
			OwnerId)
		VALUES (?,?,?,?)`,
		Name, ApiKey.String(), DEFAULT_ICON, OwnerId.Id)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

// constructDeviceFromRow converts a SQL result to device by filling out a struct.
func constructDeviceFromRow(rows *sql.Rows, err error) (*Device, error) {

	result, err := constructDevicesFromRows(rows, err)

	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0], err
	}

	return nil, errors.New("No device supplied")
}

// constructDevicesFromRows constructs a series of devices
func constructDevicesFromRows(rows *sql.Rows, err error) ([]*Device, error) {
	out := []*Device{}

	if err != nil {
		return out, err
	}

	// defensive programming
	if rows == nil {
		return out, ERR_INVALID_PTR
	}

	defer rows.Close()
	for rows.Next() {
		u := new(Device)
		err := rows.Scan(
			&u.Id,
			&u.Name,
			&u.ApiKey,
			&u.Enabled,
			&u.Icon_PngB64,
			&u.Shortname,
			&u.Superdevice,
			&u.OwnerId)

		out = append(out, u)

		if(err != nil) {
			return out, err
		}
	}

	return out, nil
}

func (userdb *UserDatabase) ReadDevicesForUserId(Id int64) ([]*Device, error) {
	rows, err := userdb.db.Query("SELECT * FROM Device WHERE OwnerId = ?", Id)

	return constructDevicesFromRows(rows, err)
}

// ReadDeviceById selects the device with the given id from the database, returning nil if none can be found
func (userdb *UserDatabase) ReadDeviceById(Id int64) (*Device, error) {
	rows, err := userdb.db.Query("SELECT * FROM Device WHERE Id = ? LIMIT 1", Id)
	return constructDeviceFromRow(rows, err)

}

// ReadDeviceByApiKey reads a device by an api key and returns it, it will be
// nil if an error was encountered and error will be set.
func (userdb *UserDatabase) ReadDeviceByApiKey(Key string) (*Device, error) {
	rows, err := userdb.db.Query("SELECT * FROM Device WHERE ApiKey = ? LIMIT 1", Key)
	return constructDeviceFromRow(rows, err)
}

// UpdateDevice updates the given device in the database with all fields in the
// struct.
func (userdb *UserDatabase) UpdateDevice(device *Device) (error) {
	if device == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.db.Exec(`UPDATE Device SET
			Name = ?, ApiKey = ?, Enabled = ?,
			Icon_PngB64 = ?, Shortname = ?, Superdevice = ?,
			OwnerId = ? WHERE Id = ?;`,
			device.Name,
			device.ApiKey,
			device.Enabled,
			device.Icon_PngB64,
			device.Shortname,
			device.Superdevice,
			device.OwnerId,
			device.Id)

	return err
}

// DeleteDevice removes a device from the system.
func (userdb *UserDatabase) DeleteDevice(Id int64) (error) {
	_, err := userdb.db.Exec(`DELETE FROM Device WHERE Id = ?;`, Id );
	return err
}

// CreateStream creates a new stream for a given device with the given name, schema and default values.
func (userdb *UserDatabase) CreateStream(Name, Schema_Json, Defaults_Json string, owner *Device) (int64, error) {
	if owner == nil {
		return -1, ERR_INVALID_PTR
	}

	res, err := userdb.db.Exec(`INSERT INTO Stream
		(	Name,
			Schema_Json,
			Defaults_Json,
			OwnerId) VALUES (?,?,?,?);`,
			Name, Schema_Json, Defaults_Json, owner.Id)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}


// constructStreamsFromRows converts a rows statement to an array of streams
func constructStreamsFromRows(rows *sql.Rows) ([]*Stream, error) {
	out := []*Stream{}

	// defensive programming
	if rows == nil {
		return out, ERR_INVALID_PTR
	}

	for rows.Next() {
		u := new(Stream)
		err := rows.Scan(
			&u.Id,
			&u.Name,
			&u.Active,
			&u.Public,
			&u.Schema_Json,
			&u.Defaults_Json,
			&u.OwnerId)

		out = append(out, u)

		if(err != nil) {
			return out, err
		}
	}

	return out, nil
}


// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *UserDatabase) ReadStreamById(id int64) (*Stream, error) {
	rows, err := userdb.db.Query("SELECT * FROM Stream WHERE Id = ? LIMIT 1", id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	streams, err := constructStreamsFromRows(rows)

	if(len(streams) != 1) {
		return nil, errors.New("Wrong number of streams returned")
	}

	return streams[0], nil
}

func (userdb *UserDatabase) ReadStreamsByDevice(device *Device) ([]*Stream, error) {
	rows, err := userdb.db.Query("SELECT * FROM Stream WHERE OwnerId = ? LIMIT 1", device.Id)

	if err != nil {
		return nil, err
	}

	return constructStreamsFromRows(rows)
}

// UpdateStream updates the stream with the given ID with the provided data
// replacing all prior contents.
func (userdb *UserDatabase) UpdateStream(stream *Stream) (error) {
	if stream == nil {
		return ERR_INVALID_PTR
	}


	_, err := userdb.db.Exec(`UPDATE Stream SET
		Name = ?,
		Active = ?,
		Public = ?,
		Schema_Json = ?,
		Defaults_Json = ?,
		OwnerId = ? WHERE Id = ?`,
		stream.Name,
		stream.Active,
		stream.Public,
		stream.Schema_Json,
		stream.Defaults_Json,
		stream.OwnerId,
		stream.Id)

	return err
}

// DeleteStream removes a stream from the database
func (userdb *UserDatabase) DeleteStream(Id int64) (error) {
	_, err := userdb.db.Exec(`DELETE FROM Stream WHERE Id = ?;`, Id );
	return err
}
