/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"config"

	"github.com/nu7hatch/gouuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidPassword = errors.New("Incorrect Password")

// calcHash calculates the user hash for the given password, salt and hashing
// scheme
func calcHash(password, salt, scheme string) (string, error) {
	saltedpass := password + salt

	switch scheme {
	// We switch over hashes here so if we need to upgrade in the future,
	// it is easy.
	case "SHA512":
		hasher := sha512.New()
		hasher.Write([]byte(saltedpass))
		return hex.EncodeToString(hasher.Sum(nil)), nil
	case "bcrypt":
		bs, err := bcrypt.GenerateFromPassword([]byte(saltedpass), bcrypt.DefaultCost)
		return string(bs), err
	default:
		return "", fmt.Errorf("Unrecognized password hash type '%s'", scheme)
	}
}

// HashPassword receives a plaintext password and returns the password, salt and type.
func HashPassword(password string) (string, string, string, error) {
	if password == "" {
		return "", "", "", errors.New("Empty Password")
	}
	salt, err := uuid.NewV4()
	if err != nil {
		return "", "", "", err
	}
	saltstr := salt.String()

	scheme := config.Get().PasswordHash
	hashed, err := calcHash(password, saltstr, scheme)

	return hashed, saltstr, scheme, err
}

// CheckPassword checks to see if the password matches the stored version
func CheckPassword(password, hashed, salt, scheme string) error {
	saltedpass := password + salt
	switch scheme {
	// We switch over hashes here so if we need to upgrade in the future,
	// it is easy.
	case "SHA512":
		h, err := calcHash(password, salt, scheme)
		if err != nil {
			return err
		}
		if h != hashed {
			return ErrInvalidPassword
		}
		return nil
	case "bcrypt":
		// Bcrypt has additional difficulty that is checked
		return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(saltedpass))
	default:
		return fmt.Errorf("Unrecognized password hash type '%s'", scheme)
	}
}

// UpgradePassword returns true if the password hashing should be upgraded
func UpgradePassword(hashed, salt, scheme string) bool {
	pscheme := config.Get().PasswordHash

	if scheme != pscheme {
		return true
	}
	if scheme == "bcrypt" {
		// Upgrade bcrypt cost if the password used an outdated cost
		c, err := bcrypt.Cost([]byte(hashed))
		if err != nil {
			return true
		}
		if c < bcrypt.DefaultCost {
			return true
		}
	}

	return false
}
