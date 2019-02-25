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
-- Groups are the underlying container for streams and access control
CREATE TABLE groups (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,
	name VARCHAR NOT NULL,
	fullname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	icon VARCHAR DEFAULT '',
	owner VARCHAR(36),

	CONSTRAINT groupowner
		FOREIGN KEY(owner) 
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE INDEX groupowners ON groups (owner,name);

-- A user is a group with an additional password
CREATE TABLE users (
	id VARCHAR(36) PRIMARY KEY NOT NULL,
	password VARCHAR NOT NULL,

	CONSTRAINT usergroup
		FOREIGN KEY (id) 
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
		
);

-- An apikey is a group with an API key.
-- We explictly link groups with API keys because that is the main use:
-- to allow something to add data to streams in its own group.
-- We also permit api keys without a linked group to behave as a raw accessor,
-- without ownership of anything
CREATE TABLE apikeys (
	apikey VARCHAR UNIQUE PRIMARY KEY NOT NULL,
	groupid VARCHAR(36),
	
	
	CONSTRAINT keygroup
		FOREIGN KEY (groupid) 
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
CREATE INDEX apikeyidx ON apikeys (apikey);

CREATE TABLE streams (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,
	name VARCHAR NOT NULL,
	nickname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	icon VARCHAR DEFAULT '',
	owner VARCHAR(36),

	schema VARCHAR NOT NULL,
	ephemeral BOOLEAN DEFAULT FALSE, -- an ephemeral stream does not save data, it just sends it through messaging
	action BOOLEAN DEFAULT FALSE, -- whether the stream permits action (intervention). 0.3 called these downlinks.
	external VARCHAR,	-- If NULL, not virtual. If not null, gives the handler url for the stream

	UNIQUE(owner,name),
	CONSTRAINT groupowner
		FOREIGN KEY(owner) 
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE	
);

CREATE INDEX streamIndex ON streams (owner,name);


----------------------------------------------------------------------
-- Permissions
----------------------------------------------------------------------

-- These are all identical tables, giving (target,actor) pairs specifying
-- if the given target can have the action done to it by the actor group


CREATE TABLE stream_permissions (
	target VARCHAR(36) NOT NULL,
	actor VARCHAR(36) NOT NULL,

	streamread BOOLEAN DEFAULT FALSE,
	streamwrite BOOLEAN DEFAULT FALSE,
	streamdelete BOOLEAN DEFAULT FALSE,

	dataread BOOLEAN DEFAULT FALSE,
	datawrite BOOLEAN DEFAULT FALSE,
	dataremove BOOLEAN DEFAULT FALSE,
	actionwrite BOOLEAN DEFAULT FALSE, -- actions are append-only

	UNIQUE(target,actor),
	PRIMARY KEY (target,actor),

	CONSTRAINT targetisstream
		FOREIGN KEY(target) 
		REFERENCES streams(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE,
	CONSTRAINT actorisgroup
		FOREIGN KEY(actor) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE
);

CREATE TABLE group_permissions (
	target VARCHAR(36) NOT NULL,
	actor VARCHAR(36) NOT NULL,

	-- group actions prefixed with g, so we can use same column names for streams
	groupread BOOLEAN DEFAULT FALSE,
	groupwrite BOOLEAN DEFAULT FALSE,
	groupdelete BOOLEAN DEFAULT FALSE,

	addstream BOOLEAN DEFAULT FALSE,
	addchild BOOLEAN DEFAULT FALSE,

	liststreams BOOLEAN DEFAULT FALSE,
	listchildren BOOLEAN DEFAULT FALSE,
	listshared BOOLEAN DEFAULT FALSE, 	

	-- child & shared stream permissions
	-- shared streams will intersect with the streams' actual permissions

	streamread BOOLEAN DEFAULT FALSE,
	streamwrite BOOLEAN DEFAULT FALSE,
	streamdelete BOOLEAN DEFAULT FALSE,

	dataread BOOLEAN DEFAULT FALSE,
	datawrite BOOLEAN DEFAULT FALSE,
	dataremove BOOLEAN DEFAULT FALSE,
	actionwrite BOOLEAN DEFAULT FALSE, -- actions are append-only

	UNIQUE(target,actor),
	PRIMARY KEY (target,actor),

	CONSTRAINT targetisgroup
		FOREIGN KEY(target) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE,
	CONSTRAINT actorisgroup
		FOREIGN KEY(actor) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE
);


----------------------------------------------------------------------
-- APIKey Permissions
----------------------------------------------------------------------

-- A key belongs to a user - by default it doesn't have any permissions,
-- but it can be given permissions that are intersected with the user's permissions.
-- This is necessary because things can be shared with a user that the user can't add permissions for.
-- These tables allow the user to give access to anything it can read to a key


CREATE TABLE apikey_stream_permissions (
	target VARCHAR(36) NOT NULL,
	actor VARCHAR(36) NOT NULL,

	read BOOLEAN DEFAULT FALSE,
	write BOOLEAN DEFAULT FALSE,
	remove BOOLEAN DEFAULT FALSE,

	rdata BOOLEAN DEFAULT FALSE,
	wdata BOOLEAN DEFAULT FALSE,
	remdata BOOLEAN DEFAULT FALSE,
	waction BOOLEAN DEFAULT FALSE, -- actions are append-only

	UNIQUE(target,actor),
	PRIMARY KEY (target,actor),

	CONSTRAINT targetisstream
		FOREIGN KEY(target) 
		REFERENCES streams(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE,
	CONSTRAINT actorisgroup
		FOREIGN KEY(actor) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE
);

CREATE TABLE apikey_group_permissions (
	target VARCHAR(36) NOT NULL,
	actor VARCHAR(36) NOT NULL,

	-- group actions prefixed with g, so we can use same column names for streams
	gread BOOLEAN DEFAULT FALSE,
	gwrite BOOLEAN DEFAULT FALSE,
	gupdate BOOLEAN DEFAULT FALSE,

	listchildren BOOLEAN DEFAULT FALSE, -- can list the direct children of the group
	listall BOOLEAN DEFAULT FALSE, 		-- can list streams shared with the group
	addstream BOOLEAN DEFAULT FALSE,

	-- child stream permissions

	read BOOLEAN DEFAULT FALSE,
	write BOOLEAN DEFAULT FALSE,
	remove BOOLEAN DEFAULT FALSE,

	rdata BOOLEAN DEFAULT FALSE,
	wdata BOOLEAN DEFAULT FALSE,
	remdata BOOLEAN DEFAULT FALSE,
	waction BOOLEAN DEFAULT FALSE,

	UNIQUE(target,actor),
	PRIMARY KEY (target,actor),

	CONSTRAINT targetisgroup
		FOREIGN KEY(target) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE,
	CONSTRAINT actorisgroup
		FOREIGN KEY(actor) 
		REFERENCES groups(id) 
		ON DELETE CASCADE
		ON UPDATE CASCADE
);



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

	if sqltype == "sqlite3" {
		_, err = db.Exec(sqliteAddonSchema)
		if err != nil {
			return err
		}
	}

	return db.Close()
}
