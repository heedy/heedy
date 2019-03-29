package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/heedy/heedy/backend/assets"

	"github.com/jmoiron/sqlx"

	// Make sure we include sqlite support
	_ "github.com/mattn/go-sqlite3"
)

// Open opens the database given assets.
func Open(a *assets.Assets) (*AdminDB, error) {

	if a.Config.SQL == nil {
		return nil, errors.New("No SQL connection string specified")
	}

	// Split the sql string into database type and connection string
	sqlInfo := strings.SplitAfterN(*a.Config.SQL, "://", 2)
	if len(sqlInfo) != 2 {
		return nil, errors.New("Invalid sql connection string")
	}
	sqltype := strings.TrimSuffix(sqlInfo[0], "://")

	if sqltype != "sqlite3" {
		return nil, fmt.Errorf("Database type '%s' not supported", sqltype)
	}

	// We use the sql as location of our sqlite database
	sqlpath := a.Abs(sqlInfo[1])

	db, err := sqlx.Open(sqltype, sqlpath)
	if err != nil {
		return nil, err
	}

	adminDB := &AdminDB{a: a}
	adminDB.SqlxCache.InitCache(db)
	return adminDB, nil
}
