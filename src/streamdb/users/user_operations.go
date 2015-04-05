// Package users provides an API for managing user information.
package users



// CreateUser creates a user given the user's credentials.
// If a user already exists with the given credentials, an error is thrown.
func (userdb *UserDatabase) CreateUser(Name, Email, Password string) error {
	if userdb.IsReadOnly() {
		return READONLY_ERR
	}

	existing, err := userdb.ReadByNameOrEmail(Name, Email)

	// Check for existance of user to provide helpful notices
	switch {
		case err != nil:
			return  err
		case existing.Email == Email:
			return ERR_EMAIL_EXISTS
		case existing.Name == Name:
			return ERR_USERNAME_EXISTS
	}

	dbpass, salt, hashtype := UpgradePassword(Password)

	_, err = userdb.Exec(`INSERT INTO Users (
	    Name,
	    Email,
	    Password,
	    PasswordSalt,
	    PasswordHashScheme) VALUES (?,?,?,?,?);`,
		Name,
		Email,
		dbpass,
		salt,
		hashtype)

	return err
}


// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadByNameOrEmail(Name, Email string) (*User, error) {
	var exists User

	err := userdb.Get(&exists, "SELECT * FROM Users WHERE Name = ? OR Email = ? LIMIT 1;", Name, Email)

	return &exists, err
}

// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadUserByEmail(Email string) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Email = ? LIMIT 1;", Email)

	return &user, err
}

// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *UserDatabase) ReadUserByName(Name string) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Name = ? LIMIT 1;", Name)

	return &user, err
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *UserDatabase) ReadUserById(UserId int64) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE UserId = ? LIMIT 1;", UserId)

	return &user, err
}

func (userdb *UserDatabase) ReadAllUsers() ([]User, error) {
	var users []User

	err := userdb.Select(&users, "SELECT * FROM Users")

	return users, err
}

func (userdb *UserDatabase) ReadStreamOwner(StreamId int64) (*User, error) {
	var user User

	err := userdb.Get(&user, `SELECT u.*
	                              FROM Users u, Stream s, Device d
	                              WHERE s.StreamId = ?
	                                AND d.DeviceId = s.DeviceId
	                                AND u.UserId = d.UserId
	                              LIMIT 1;`, StreamId)

	return &user, err
}

// UpdateUser updates the user with the given id in the database using the
// information provided in the user struct.
func (userdb *UserDatabase) UpdateUser(user *User) error {
	if userdb.IsReadOnly() {
		return READONLY_ERR
	}

	if user == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.Exec(`UPDATE Users SET
	                Name=?, Email=?, Password=?, PasswordSalt=?, PasswordHashScheme=?,
	                Admin=?, UploadLimit_Items=?,
	                ProcessingLimit_S=?, StorageLimit_Gb=? WHERE UserId = ?`,
		user.Name,
		user.Email,
		user.Password,
		user.PasswordSalt,
		user.PasswordHashScheme,
		user.Admin,
		user.UploadLimit_Items,
		user.ProcessingLimit_S,
		user.StorageLimit_Gb,
		user.UserId)
	
	return err
}

// DeleteUser removes a user from the database
func (userdb *UserDatabase) DeleteUser(UserId int64) error {
	if userdb.IsReadOnly() {
		return READONLY_ERR
	}

	_, err := userdb.Exec(`DELETE FROM Users WHERE UserId = ?;`, UserId)
	return err
}
