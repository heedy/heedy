package users

import (
	"github.com/nu7hatch/gouuid"
	"crypto/sha512"
	"encoding/hex"
    "log"
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

// Receives a plaintext password and returns the password, salt and type.
func UpgradePassword(password string) (string, string, string){
	salt, _ := uuid.NewV4()
	saltstr := salt.String()

	scheme := "SHA512"
	hashed := calcHash(password, saltstr, scheme)

	return hashed, saltstr, "SHA512"
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
