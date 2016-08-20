package dbutil

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

// OpenDatabase opens an alread-created database
func OpenDatabase(dbtype, uri string) (*sqlx.DB, error) {
	log.Debugf("Opening %s database at %s", dbtype, uri)
	db, err := sqlx.Open(dbtype, uri)
	if err != nil {
		return nil, err
	}

	// Now let's query the database version to make sure we can read it!
	var version string
	row := db.QueryRowx("SELECT Value FROM connectordbmeta WHERE Key='DBVersion';")
	if row.Err() != nil {
		return nil, row.Err()
	}
	err = row.Scan(&version)
	if err != nil {
		return nil, err
	}
	if version != "20160820" {
		return nil, errors.New("The existing database is incompatible with this version of ConnectorDB")
	}
	return db, nil
}
