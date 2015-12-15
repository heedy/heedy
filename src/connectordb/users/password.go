/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"crypto/sha512"
	"encoding/hex"

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

// Receives a plaintext password and returns the password, salt and type.
func UpgradePassword(password string) (string, string, string) {
	salt, _ := uuid.NewV4()
	saltstr := salt.String()

	scheme := "SHA512"
	hashed := calcHash(password, saltstr, scheme)

	return hashed, saltstr, "SHA512"
}
