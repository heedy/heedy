package registry

import (
	"database/sql"
	// Make sure we include sqlite support

	"os"
	"path/filepath"
	"time"

	"github.com/blang/semver/v4"
	"github.com/jmoiron/sqlx"

	// The regsitry is _always_ an sqlite database
	_ "github.com/mattn/go-sqlite3"
)

// Version is the registry file version
var Version = semver.MustParse("1.0.0")

var schema = `

CREATE TABLE plugins (
	name VARCHAR UNIQUE PRIMARY KEY NOT NULL, 	-- Name of the plugin
	icon VARCHAR,		 		-- plugin icon
	fullname VARCHAR NOT NULL, 	-- full name of plugin
	description VARCHAR,		-- Plugin Description
	version VARCHAR NOT NULL,	-- Semver version
	heedy_version VARCHAR,		-- Semver compatible heedy versions
	webpage VARCHAR,			-- URL to repository
	release_url VARCHAR,        -- URL to zip file of release
	python BOOLEAN,				-- whether the plugin requires python
	license VARCHAR,			-- Name of the license
	stars INTEGER,				-- Number of github stars
	timestamp INTEGER,			-- Timestamp when this entry was last updated (unix)

	UNIQUE(webpage),
	UNIQUE(release_url)
);

CREATE INDEX pluginstars ON plugins(stars);
CREATE INDEX plugintime ON plugins(timestamp);
CREATE INDEX pluginurl ON plugins(webpage);
CREATE INDEX pluginrelease ON plugins(release_url);

-- The metadata table holds the following keys:
-- > "registry", semver version of registry
-- > "updated", timestamp that the database was last updated, RFC3339
-- > "heedy", semver version of most recent heedy (optional)
CREATE TABLE metadata (
	k VARCHAR PRIMARY KEY,
	v VARCHAR
);

`

// Registry holds info about the... registry!
type Registry struct {
	db *sqlx.DB

	heedyVersion    *semver.Version
	RegistryVersion semver.Version
	Updated         time.Time
}

// Plugin holds info about the plugin from the registry
type Plugin struct {
	Name         string
	Icon         string
	FullName     string
	Description  string
	Version      semver.Version
	heedyVersion semver.Version
	Webpage      string
	ReleaseURL   string
	Python       bool
	License      string
	Stars        int
	Timestamp    time.Time
}

// Create generates a new regsitry database file
func Create(filename string) (*Registry, error) {

	sqlpath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	// Create any necessary directories
	sqlfolder := filepath.Dir(sqlpath)
	if err := os.MkdirAll(sqlfolder, 0750); err != nil {
		return nil, err
	}

	db, err := sqlx.Open("sqlite3", sqlpath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Next, we insert the registry version, and current timestamp into
	_, err = db.Exec("INSERT INTO metadata (k,v) VALUES ('registry',?);", Version)

	if err != nil {
		return nil, err
	}
	r := &Registry{db: db, RegistryVersion: Version}
	t, err := r.UpdateRegistry()
	r.Updated = t
	return r, err

}

// Open opens a registry file
func Open(filename string) (*Registry, error) {
	sqlpath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("sqlite3", sqlpath)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		db: db,
	}
	v, err := r.Get("updated")
	if err != nil {
		db.Close()
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		db.Close()
		return nil, err
	}
	r.Updated = t

	if v, err = r.Get("registry"); err != nil {
		db.Close()
		return nil, err
	}
	ver, err := semver.Parse(v)
	if err != nil {
		db.Close()
		return nil, err
	}
	r.RegistryVersion = ver

	// And finally, check if there is a heedy version specified
	if v, err = r.Get("registry"); err != nil {
		db.Close()
		return nil, err
	}
	if v != "" {
		heedyv, err := semver.New(v)
		if err != nil {
			db.Close()
			return nil, err
		}
		r.heedyVersion = heedyv
	}

	return r, err
}

// UpdateRegistry updates the registry timestamp
func (r *Registry) UpdateRegistry() (time.Time, error) {
	t := time.Now()
	err := r.Set("updated", t.Format(time.RFC3339))
	return t, err
}

// Get gets the given key. If none is set, returns empty string
func (r *Registry) Get(key string) (string, error) {
	var val string
	err := r.db.Get(&val, "SELECT v FROM metadata WHERE k=?", key)
	if err == sql.ErrNoRows {
		return "", nil // Empty string if no result
	}
	return val, err

}

// Set sets the given key
func (r *Registry) Set(key string, value string) error {
	if value == "" {
		_, err := r.db.Exec("DELETE FROM metadata WHERE key=?", key)
		return err
	}
	_, err := r.db.Exec("INSERT OR REPLACE INTO metadata (k,v) VALUES (?,?);", key, value)
	return err
}

// Close closes the underlying database
func (r *Registry) Close() error {
	return r.db.Close()
}
