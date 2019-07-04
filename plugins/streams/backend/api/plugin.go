package api

import (
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/server"
)


// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, curversion int) error {
	if curversion != 0 {
		return errors.New("Streams database version too new")
	}
	return CreateSQLData(db.DB)
}

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when streams are compiled directly into the main heedy executable.
func init() {
	// Add the main handler to the server's builtin routes. The builtin heedy.conf will refer
	// to builtin://streams to access the handler
	server.BuiltinRoutes["streams"] = Handler
	
	// Register the sql updater, so that the tables for streams are automatically created
	// and updated on database open
	database.RegisterPlugin("streams", SQLVersion, SQLUpdater)
}
