package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/connectordb/connectordb/assets"

	// Make sure we include sqlite support
	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE user (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	nickname TEXT,
	icon TEXT
);

CREATE INDEX usernames ON user (name);
`

// Create sets up a new CDB instance
func Create(a *assets.Assets) error {

	if a.Config.SQL == nil {
		return errors.New("Configuration does not specify an sql database")
	}

	// Split the sql string into database type and connection string
	sqlInfo := strings.SplitAfterN(*a.Config.SQL, "://", 2)
	if len(sqlInfo) != 2 {
		return errors.New("Invalid sql connection string")
	}
	sqltype := strings.TrimSuffix(sqlInfo[0], "://")

	if sqltype != "sqlite3" {
		return fmt.Errorf("Database type '%s' not supported", sqltype)
	}

	// We use the sql as location of our sqlite database
	sqlpath := a.Abs(sqlInfo[1])

	// Create any necessary directories
	sqlfolder := filepath.Dir(sqlpath)
	if err := os.MkdirAll(sqlfolder, 0750); err != nil {
		return err
	}

	db, err := sql.Open(sqltype, sqlpath)
	if err != nil {
		return err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	return db.Close()
}
