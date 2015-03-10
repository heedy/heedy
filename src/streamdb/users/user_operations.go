// Package users provides an API for managing user information.
package users

import ("time"
    "database/sql"
    "errors"
    "crypto/sha512"
    "encoding/hex"
    "log"
    "github.com/nu7hatch/gouuid"
)

// calcHash calculates the user hash for the given password, salt and hashing
// scheme
func calcHash(password, salt, scheme string) string {
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
	if usr != nil {
		return -1, ERR_EMAIL_EXISTS
	}

	usr, err = userdb.ReadUserByName(Name)
	if usr != nil {
		return -1, ERR_USERNAME_EXISTS
	}

	PasswordSalt, _ := uuid.NewV4()
	dbpass := calcHash(Password, PasswordSalt.String(), DEFAULT_PASSWORD_HASH)

	// Note that golang uses utf8 strings converted to bytes first, so the hashes
	// may not match up with hash generators found online!
	//log.Print("passwordtest ", saltedpass, []byte(saltedpass), dbpass)

	res, err := userdb.Db.Exec(CREATE_USER_STMT,
		Name,
		Email,
		dbpass,
		PasswordSalt.String(),
		DEFAULT_PASSWORD_HASH,
        int64(time.Now().Unix())) // current time is creation time

	if err != nil {
		return -1, err
	}

    log.Printf("Created user %v", Name)


	return res.LastInsertId()
}

// constructUserFromRow converts a sql.Rows object to a single user
func constructUserFromRow(rows *sql.Rows, err error) (*User, error) {
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

func constructUsersFromRows(rows *sql.Rows) ([]*User, error) {
	out := []*User{}

	if rows == nil {
		return out, ERR_INVALID_PTR
	}

	for rows.Next() {
		u := new(User)
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
			&u.StorageLimit_Gb,
            &u.CreateTime,
            &u.ModifyTime,
            &u.UserGroup)

		if err != nil {
			return out, err
		}

		out = append(out, u)
	}

	return out, nil
}

// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadUserByEmail(Email string) (*User, error) {
	rows, err := userdb.Db.Query(SELECT_USER_BY_EMAIL_STMT, Email)
	return constructUserFromRow(rows, err)
}


// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *UserDatabase) ReadUserByName(Name string) (*User, error) {
	rows, err := userdb.Db.Query(SELECT_USER_BY_NAME_STMT, Name)
	return constructUserFromRow(rows, err)
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *UserDatabase) ReadUserById(Id int64) (*User, error) {
	rows, err := userdb.Db.Query(SELECT_USER_BY_ID_STMT, Id)
	return constructUserFromRow(rows, err)
}

func (userdb *UserDatabase) ReadAllUsers() ([]*User, error) {
	rows, err := userdb.Db.Query(SELECT_ALL_USERS_STMT)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return constructUsersFromRows(rows)
}



func (userdb *UserDatabase) ReadStreamOwner(StreamId int64) (*User, error) {
    rows, err := userdb.Db.Query(SELECT_OWNER_OF_STREAM_BY_ID_STMT, StreamId)

    return constructUserFromRow(rows, err)
}


// UpdateUser updates the user with the given id in the database using the
// information provided in the user struct.
func (userdb *UserDatabase) UpdateUser(user *User) error {

	if user == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.Db.Exec( UPDATE_USER_STMT,
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
        user.CreateTime,
        int64(time.Now().Unix()),        // user.ModifyTime,
        user.UserGroup,
		user.Id)
	return err
}

// DeleteUser removes a user from the database
func (userdb *UserDatabase) DeleteUser(id int64) error {
	_, err := userdb.Db.Exec(DELETE_USER_BY_ID_STMT, id)
	return err
}
