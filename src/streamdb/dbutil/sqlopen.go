package dbutil

import (
	"database/sql"
    "log"
	"strings"
	//The blank imports are used to automatically register the database handlers
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

    )

type DRIVERSTR string

const (
	SQLITE3  DRIVERSTR = "sqlite3"
	POSTGRES DRIVERSTR = "postgres"
)

const (
	SQLITE_PREFIX   = "sqlite://"
	POSTGRES_PREFIX = "postgres://"
)

func (d DRIVERSTR) String() string {
    return string(d)
}

// Checks if a URI is sqlite
func UriIsSqlite(sqluri string) bool {
	return strings.HasSuffix(sqluri, ".db") ||
		strings.HasSuffix(sqluri, ".sqlite") ||
		strings.HasSuffix(sqluri, ".sqlite3") ||
		strings.HasPrefix(sqluri, SQLITE_PREFIX)
}

// Strips the leading sqlite:// from a sqlite path
func SqliteURIToPath(sqluri string) string {
	//The sqlite driver doesn't like starting with sqlite://
	if strings.HasPrefix(sqluri, SQLITE_PREFIX) {
		sqluri = sqluri[len(SQLITE_PREFIX):]
	}

	return sqluri
}

// Gets the conversion script for the given database.
func OpenSqlDatabase(sqluri string) (*sql.DB, DRIVERSTR, error) {
	var err error

    sqltype := POSTGRES //The default is postgres.

	//First, we check if the user wants to use sqlite or postgres. If the url given
	//has the hallmarks of a file or sqlite database, then set that as the database type
	switch {
	// TODO just check if this is a file
	case UriIsSqlite(sqluri):
		sqltype = SQLITE3
		sqluri = SqliteURIToPath(sqluri)
		break
	case strings.HasPrefix(sqluri, POSTGRES_PREFIX):
		sqltype = POSTGRES
		sqluri = sqluri[len(POSTGRES_PREFIX):]
	default:
		log.Printf("Warning, database type was found, defaulting to %v", sqltype)
	}

	log.Printf("Opening %v database with cxn string: %v", sqltype, sqluri)

	sqldb, err := sql.Open(sqltype.String(), sqluri)

	if err != nil {
		log.Printf("Open failed\n")
		return sqldb, sqltype, err
	}

	err = sqldb.Ping()
	return sqldb, sqltype, err
}

// Gets the streamdb database version
func GetDatabaseVersion(db *sql.DB, dbtype DRIVERSTR) string{
	version := "00000000"

	var mixin SqlxMixin
	mixin.InitSqlxMixin(db, dbtype.String())

	err := mixin.Get(&version, "SELECT Value FROM StreamdbMeta WHERE Key = 'DBVersion'")

	if err != nil {
		version = "00000000"
	}

	return version
}
