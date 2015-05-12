package dbutil

import (
	"database/sql"
	"streamdb/config"
	"strings"

	log "github.com/Sirupsen/logrus"
	//The blank imports are used to automatically register the database handlers
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// The driver strings for the given database types as needed by the sql package
	SQLITE3  string = config.Sqlite
	POSTGRES string = config.Postgres
)

const (
	// The URI prefixes for the given drivers
	sqlitePrefix   = "sqlite://"
	postgresPrefix = "postgres://"
)

// Checks if a URI is sqlite
func UriIsSqlite(sqluri string) bool {
	return strings.HasSuffix(sqluri, ".db") ||
		strings.HasSuffix(sqluri, ".sqlite") ||
		strings.HasSuffix(sqluri, ".sqlite3") ||
		strings.HasPrefix(sqluri, sqlitePrefix)
}

// Strips the leading sqlite:// from a sqlite path
func SqliteURIToPath(sqluri string) string {
	//The sqlite driver doesn't like starting with sqlite://
	if strings.HasPrefix(sqluri, sqlitePrefix) {
		sqluri = sqluri[len(sqlitePrefix):]
	}

	return sqluri
}

// From a connection string, gets the cleaned connection path and database type
func ProcessConnectionString(connectionString string) (connector string, dbt string) {

	dbt = POSTGRES //The default is postgres.
	connector = connectionString

	//First, we check if the user wants to use sqlite or postgres. If the url given
	//has the hallmarks of a file or sqlite database, then set that as the database type
	switch {
	// TODO just check if this is a file
	case UriIsSqlite(connectionString):
		dbt = SQLITE3
		connector = SqliteURIToPath(connectionString)
		break
	case strings.HasPrefix(connectionString, postgresPrefix):
		dbt = POSTGRES
	default:
		log.Warningf("database type was found, defaulting to %v", dbt)
	}

	return connector, dbt
}

// Gets the conversion script for the given database.
func OpenSqlDatabase(connectionString string) (*sql.DB, string, error) {
	var err error

	sqluri, sqltype := ProcessConnectionString(connectionString)
	log.Debugf("Opening %v database with cxn string: %v", sqltype, sqluri)

	sqldb, err := sql.Open(sqltype, sqluri)

	return sqldb, sqltype, err
}

// Gets the streamdb database version
func GetDatabaseVersion(db *sql.DB, dbtype string) string {
	version := defaultDbversion

	var mixin SqlxMixin
	mixin.InitSqlxMixin(db, dbtype)

	err := mixin.Get(&version, "SELECT Value FROM StreamdbMeta WHERE Key = 'DBVersion'")

	if err != nil {
		version = defaultDbversion
	}

	return version
}
