package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GenerateKey creates a random API key
func GenerateKey(length int) (string, error) {
	// Prepare the plugin API key
	apikey := make([]byte, length)
	_, err := rand.Read(apikey)
	return base64.StdEncoding.EncodeToString(apikey), err
}

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	passwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(passwd), err
}

// CheckPassword checks if the password is valid
func CheckPassword(password, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

var (
	nameValidator = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")
)

func ValidUserName(name string) error {
	if nameValidator.MatchString(name) && len(name) > 0 {
		return nil
	}
	return ErrInvalidUserName
}

// Ensures that the icon is in a valid format
func ValidIcon(icon string) error {
	if icon == "" {
		return nil
	}
	if !strings.HasPrefix(icon, "data:image/") {
		if len(icon) > 30 {
			return errors.New("bad_request: icon name can't be more than 30 characters unless it is an image")
		}
		return nil
	}
	return nil
}

// Checks whether the given group access level is OK
func ValidGroupScopes(s *ScopeArray) error {
	return nil
}

// Performs a set of tests on the result and error of a
// call to see what kind of error we should return.
func getExecError(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
