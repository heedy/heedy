package auth

import (
	"net/http"

	"github.com/connectordb/connectordb/src/database"
)

// Creator handles generation of credentials and tokens
func Creator(w http.ResponseWriter, r *http.Request) {

}

// Authenticate extracts the appropriate database from a request
func Authenticate(db *database.AdminDB, r *http.Request) (database.DB, error) {
	return db, nil
}
