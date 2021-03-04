package database

import (
	"database/sql"
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
		return nil, errors.New("No SQL app string specified")
	}

	// Split the sql string into database type and app string
	sqlInfo := strings.SplitAfterN(*a.Config.SQL, "://", 2)
	if len(sqlInfo) != 2 {
		return nil, errors.New("Invalid sql app string")
	}
	sqltype := strings.TrimSuffix(sqlInfo[0], "://")

	if sqltype != "sqlite3" {
		return nil, fmt.Errorf("Database type '%s' not supported", sqltype)
	}

	drv := sql.Drivers()
	for _, d := range drv {
		// The events module registered sqlite3_heedy to hook into database modifications,
		// that way events can be auto-dispatched withot worrying about cascade and extra stuff.
		// Plugins are also free to register _heedy versions of database drivers, so that
		// they can hook into events of their own.
		if sqltype+"_heedy" == d {
			sqltype = sqltype + "_heedy"
		}
	}

	// We use the sql as location of our sqlite database
	sqlpath := a.DataAbs(sqlInfo[1])

	db, err := sqlx.Open(sqltype, sqlpath)
	if err != nil {
		return nil, err
	}

	adminDB := &AdminDB{
		a: a,
	}
	adminDB.SqlxCache.InitCache(db)
	if a.Config.Verbose {
		adminDB.SqlxCache.Verbose = true
	}

	hversion, err := adminDB.ReadPluginDatabaseVersion("heedy")
	if err != nil {
		return nil, err
	}
	if hversion < 2 || hversion > 2 {
		return nil, errors.New("The given database is incompatible with this version of Heedy")
	}

	return adminDB, nil
}
