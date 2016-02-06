/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package dbutil

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	//The blank imports are used to automatically register the database handlers
	_ "github.com/lib/pq"
)

const (
	// The driver strings for the given database types as needed by the sql package
	POSTGRES string = "postgres"
)

const (
	// The URI prefixes for the given drivers
	postgresPrefix = "postgres://"
)

// Gets the conversion script for the given database.
func OpenSqlDatabase(sqluri string) (*sql.DB, string, error) {
	var err error
	sqltype := POSTGRES

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
